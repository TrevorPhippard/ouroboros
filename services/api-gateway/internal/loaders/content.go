package loaders

import (
	"context"
	"fmt"
	"sort"

	"api-gateway/graph/model"
	"api-gateway/internal/args"
	"api-gateway/internal/mapping"

	postpb "ouroboros/proto/generated/post"
)

// fetchPosts resolves posts by ID with one GetPostsByIds call.
func (d Deps) fetchPosts(ctx context.Context, ids []string) (map[string]*model.Post, error) {
	ids = args.UniqueIDs(ids)
	posts := make(map[string]*model.Post, len(ids))
	if len(ids) == 0 {
		return posts, nil
	}
	res, err := d.Clients.Post.GetPostsByIds(ctx, &postpb.GetPostsByIdsRequest{Ids: ids})
	if err != nil {
		return nil, fmt.Errorf("failed to fetch posts: %w", err)
	}
	if res == nil {
		return posts, nil
	}
	for _, p := range res.Posts {
		if p != nil {
			posts[p.Id] = mapping.Post(p)
		}
	}
	return posts, nil
}

// fetchPostsByAuthor groups keys by limit and issues one catalog call per
// distinct limit (normally exactly one).
func (d Deps) fetchPostsByAuthor(ctx context.Context, keys []listKey) (map[listKey][]*model.Post, error) {
	out := make(map[listKey][]*model.Post, len(keys))
	if d.Catalog == nil {
		// Nested lists degrade to empty rather than failing the whole query.
		for _, k := range keys {
			out[k] = []*model.Post{}
		}
		return out, nil
	}
	for limit, ids := range groupByLimit(keys) {
		byAuthor, err := d.Catalog.PostsByAuthors(ctx, ids, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch posts by author: %w", err)
		}
		for _, id := range ids {
			out[listKey{ID: id, Limit: limit}] = Cap(byAuthor[id], int(limit))
		}
	}
	return out, nil
}

func (d Deps) fetchComments(ctx context.Context, keys []listKey) (map[listKey][]*model.Comment, error) {
	out := make(map[listKey][]*model.Comment, len(keys))
	if d.Catalog == nil {
		for _, k := range keys {
			out[k] = []*model.Comment{}
		}
		return out, nil
	}
	for limit, ids := range groupByLimit(keys) {
		byPost, err := d.Catalog.CommentsByPosts(ctx, ids, limit)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch comments: %w", err)
		}
		for _, id := range ids {
			out[listKey{ID: id, Limit: limit}] = Cap(byPost[id], int(limit))
		}
	}
	return out, nil
}

// groupByLimit returns the distinct IDs for each limit, in a stable order.
func groupByLimit(keys []listKey) map[int32][]string {
	grouped := make(map[int32][]string)
	for _, k := range keys {
		grouped[k.Limit] = append(grouped[k.Limit], k.ID)
	}
	for limit := range grouped {
		sort.Strings(grouped[limit])
	}
	return grouped
}

// Cap enforces a limit even if a backend returns more than it was asked for,
// drops nil entries, and never returns nil (GraphQL non-null lists must
// serialize as []).
func Cap[T any](items []*T, limit int) []*T {
	out := make([]*T, 0, len(items))
	for _, it := range items {
		if it == nil {
			continue
		}
		out = append(out, it)
		if len(out) >= limit {
			break
		}
	}
	return out
}
