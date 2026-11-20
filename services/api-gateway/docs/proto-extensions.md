# Backend RPCs needed for the remaining list fields

Everything in the gateway is implemented, including batching, limits, ordering
and authorization. Four GraphQL fields can only return real data once the
downstream services expose list RPCs, which the gateway's current protobuf
clients do not have:

| GraphQL field          | Needs                                  |
|------------------------|----------------------------------------|
| `Query.posts`          | `PostService.ListPosts`                |
| `User.posts`           | `PostService.GetPostsByAuthors`        |
| `Post.comments`        | `PostService.GetCommentsByPostIds`     |
| `Query.recommendations`| `ConnectionService.GetRecommendations` |

Until then (default build) `Query.posts` and `Query.recommendations` return an
explicit "unavailable" error and `User.posts` / `Post.comments` return `[]`.

## Proposed additions

The two `Get...` RPCs are deliberately **batch** RPCs so the gateway's data
loaders can resolve a whole page of users/posts in one call.

```proto
// post.proto
service PostService {
  // Newest first, at most `limit` posts.
  rpc ListPosts(ListPostsRequest) returns (ListPostsResponse);
  // Newest first, at most `limit_per_author` posts for EACH author.
  rpc GetPostsByAuthors(GetPostsByAuthorsRequest) returns (GetPostsByAuthorsResponse);
  // Newest first, at most `limit_per_post` comments for EACH post.
  rpc GetCommentsByPostIds(GetCommentsByPostIdsRequest) returns (GetCommentsByPostIdsResponse);
}

message ListPostsRequest  { int32 limit = 1; }
message ListPostsResponse { repeated Post posts = 1; }

message GetPostsByAuthorsRequest  { repeated string author_ids = 1; int32 limit_per_author = 2; }
message AuthorPosts               { string author_id = 1; repeated Post posts = 2; }
message GetPostsByAuthorsResponse { repeated AuthorPosts authors = 1; }

message GetCommentsByPostIdsRequest  { repeated string post_ids = 1; int32 limit_per_post = 2; }
message PostComments                 { string post_id = 1; repeated Comment comments = 2; }
message GetCommentsByPostIdsResponse { repeated PostComments posts = 1; }
```

```proto
// connection.proto
service ConnectionService {
  // Accounts `user_id` might want to follow, best match first. Must exclude
  // `user_id` itself and accounts already followed (the gateway also filters).
  rpc GetRecommendations(GetRecommendationsRequest) returns (GetRecommendationsResponse);
}

message GetRecommendationsRequest  { string user_id = 1; int32 limit = 2; }
message GetRecommendationsResponse { repeated string user_ids = 1; }
```

## Turning it on

1. Add the RPCs, implement them in the services, regenerate `ouroboros/proto/generated/...`.
2. Build the gateway with the tag: `go build -tags postext ./cmd` (or `go run -tags postext ./cmd`).

The adapter is `internal/catalog/grpc.go` (wired by `catalog.FromClients`). If you
choose different RPC or field names, only that file changes.
