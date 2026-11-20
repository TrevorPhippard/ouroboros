// Package loaders provides the per-operation data loaders.
//
// A fresh Loaders is created for every GraphQL operation (see Extension), so its
// cache can never leak data between users, and every key requested while one
// level of a query is being resolved is fetched with a single downstream call.
// Resolvers use the accessor methods and never see the loaders themselves.
package loaders

import (
	"context"

	"api-gateway/graph/model"
	"api-gateway/internal/args"
	"api-gateway/internal/catalog"
	"api-gateway/internal/clients"
	"api-gateway/internal/dataloader"

	"github.com/99designs/gqlgen/graphql"
	"github.com/vektah/gqlparser/v2/ast"
)

// Deps is what the loaders need to reach the backend.
type Deps struct {
	Clients clients.Set
	// Catalog is optional. Without it, author posts and post comments load as
	// empty lists.
	Catalog catalog.PostCatalog
}

// listKey identifies "the first N items of parent X". The limit is part of the
// key because the same parent can be requested with different limits via
// GraphQL aliases.
type listKey struct {
	ID    string
	Limit int32
}

// Loaders batches and caches lookups for one GraphQL operation.
type Loaders struct {
	users         *dataloader.Loader[string, *model.User]
	posts         *dataloader.Loader[string, *model.Post]
	postsByAuthor *dataloader.Loader[listKey, []*model.Post]
	comments      *dataloader.Loader[listKey, []*model.Comment]
}

// New builds a fresh set of loaders bound to ctx. ctx must carry the caller's
// auth metadata and identity because batch functions run with it.
func New(ctx context.Context, d Deps, cfg dataloader.Config) *Loaders {
	return &Loaders{
		users:         dataloader.New(ctx, cfg, d.fetchUsers),
		posts:         dataloader.New(ctx, cfg, d.fetchPosts),
		postsByAuthor: dataloader.New(ctx, cfg, d.fetchPostsByAuthor),
		comments:      dataloader.New(ctx, cfg, d.fetchComments),
	}
}

type ctxKey struct{}

// With stores l in ctx.
func With(ctx context.Context, l *Loaders) context.Context {
	return context.WithValue(ctx, ctxKey{}, l)
}

// From returns the operation's loaders, or nil when none are installed.
func From(ctx context.Context) *Loaders {
	l, _ := ctx.Value(ctxKey{}).(*Loaders)
	return l
}

// For returns the operation's loaders. When none are installed (a resolver
// called directly, say from a test) it falls back to uncached single-use
// loaders: slower, but still correct.
func (d Deps) For(ctx context.Context) *Loaders {
	if l := From(ctx); l != nil {
		return l
	}
	return New(ctx, d, dataloader.Config{DisableCache: true})
}

// Extension is a gqlgen extension that installs fresh Loaders for every
// operation, over HTTP and WebSocket alike. Add it with srv.Use(...).
func Extension(d Deps) graphql.HandlerExtension { return extension{d: d} }

type extension struct{ d Deps }

var _ graphql.OperationInterceptor = extension{}

func (extension) ExtensionName() string                   { return "DataLoaders" }
func (extension) Validate(graphql.ExecutableSchema) error { return nil }

func (e extension) InterceptOperation(ctx context.Context, next graphql.OperationHandler) graphql.ResponseHandler {
	cfg := dataloader.Config{}
	// A subscription lives as long as its socket and re-resolves nested fields
	// for every event, so caching would serve stale users.
	if oc := graphql.GetOperationContext(ctx); oc != nil && oc.Operation != nil && oc.Operation.Operation == ast.Subscription {
		cfg.DisableCache = true
	}
	return next(With(ctx, New(ctx, e.d, cfg)))
}

// ---------------------------------------------------------------------------
// Accessors used by resolvers
// ---------------------------------------------------------------------------

// User returns one user, or nil if the user does not exist.
func (l *Loaders) User(ctx context.Context, id string) (*model.User, error) {
	ids := args.UniqueIDs([]string{id})
	if len(ids) == 0 {
		return nil, nil
	}
	return l.users.Load(ctx, ids[0])
}

// Users returns the requested users keyed by ID. Unknown users map to nil.
func (l *Loaders) Users(ctx context.Context, ids []string) (map[string]*model.User, error) {
	ids = args.UniqueIDs(ids)
	if len(ids) == 0 {
		return map[string]*model.User{}, nil
	}
	return l.users.LoadMany(ctx, ids)
}

// ClearUser evicts a user after a mutation changed it, so a later selection in
// the same operation reads the new state.
func (l *Loaders) ClearUser(id string) { l.users.Clear(id) }

// Post returns one post, or nil if it does not exist.
func (l *Loaders) Post(ctx context.Context, id string) (*model.Post, error) {
	return l.posts.Load(ctx, id)
}

// Posts returns the requested posts keyed by ID. Unknown posts map to nil.
func (l *Loaders) Posts(ctx context.Context, ids []string) (map[string]*model.Post, error) {
	return l.posts.LoadMany(ctx, args.UniqueIDs(ids))
}

// PostsBy returns an author's most recent posts (never nil).
func (l *Loaders) PostsBy(ctx context.Context, authorID string, limit int32) ([]*model.Post, error) {
	posts, err := l.postsByAuthor.Load(ctx, listKey{ID: authorID, Limit: limit})
	if posts == nil {
		posts = []*model.Post{}
	}
	return posts, err
}

// CommentsOn returns a post's most recent comments (never nil).
func (l *Loaders) CommentsOn(ctx context.Context, postID string, limit int32) ([]*model.Comment, error) {
	comments, err := l.comments.Load(ctx, listKey{ID: postID, Limit: limit})
	if comments == nil {
		comments = []*model.Comment{}
	}
	return comments, err
}
