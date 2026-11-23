// Package catalog defines the read capabilities that go beyond what the Post
// and Connection services expose today: listing posts, a user's posts, a post's
// comments, and follow recommendations.
//
// They are interfaces in the gateway's own types so batching, limits and
// authorization are implemented and tested independently of the protobuf
// definitions. FromClients wires the gRPC implementation when the binary is
// built with `-tags postext` (see docs/proto-extensions.md); otherwise the
// capabilities are absent and the schema fields degrade explicitly.
package catalog

import (
	"context"
	"errors"

	"api-gateway/graph/model"
)

// ErrUnavailable is returned when a field depends on a capability this
// deployment has not wired up.
var ErrUnavailable = errors.New("not available: the downstream service does not expose this capability")

// PostCatalog is the read side of the Post service beyond fetch-by-ID.
//
// Implementations must return at most `limit` entries per key, newest first,
// omit (or return empty slices for) keys with no entries, and be safe for
// concurrent use.
type PostCatalog interface {
	// ListPosts returns the most recent posts across all authors.
	ListPosts(ctx context.Context, limit int32) ([]*model.Post, error)
	// PostsByAuthors returns each author's most recent posts, keyed by author ID.
	PostsByAuthors(ctx context.Context, authorIDs []string, limit int32) (map[string][]*model.Post, error)
	// CommentsByPosts returns each post's most recent comments, keyed by post ID.
	CommentsByPosts(ctx context.Context, postIDs []string, limit int32) (map[string][]*model.Comment, error)
}

// Recommender suggests accounts for a user to follow.
type Recommender interface {
	// RecommendedUserIDs returns up to limit user IDs, best match first. It
	// must not include userID itself or accounts userID already follows.
	RecommendedUserIDs(ctx context.Context, userID string, limit int32) ([]string, error)
}
