# API Gateway

GraphQL gateway for the Ouroboros auth, connection, feed, notification, post, and profile gRPC services.

## Prerequisites

- Go version `1.25.8` (as declared in `go.mod`).
- The parent Ouroboros repository/workspace containing `ouroboros/proto/generated/{auth,connection,feed,notification,post,profile}`. Those generated protobuf packages are not included in this gateway-only archive.
- The auth service must issue HS256 JWTs for the included HTTP auth middleware, with an `exp` claim and a user identifier in one of `sub`, `user_id`, `userId`, `userID`, or `id`.

## Configuration

The following environment variables are supported:

| Variable | Default | Purpose |
|---|---|---|
| `PORT` | `4000` | HTTP listen port |
| `AUTH_SERVICE_URL` | `localhost:50053` | Auth gRPC endpoint |
| `CONNECTION_SERVICE_URL` | `localhost:50054` | Connection gRPC endpoint |
| `FEED_SERVICE_URL` | `localhost:50055` | Feed gRPC endpoint |
| `NOTIFICATION_SERVICE_URL` | `localhost:50056` | Notification gRPC endpoint |
| `POST_SERVICE_URL` | `localhost:50057` | Post gRPC endpoint |
| `PROFILE_SERVICE_URL` | `localhost:50058` | Profile gRPC endpoint |
| `JWT_SECRET` | unset | Shared HS256 verification secret; protected operations fail closed if unset. There is no default |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:3000,http://127.0.0.1:3000` | Comma-separated browser origins |
| `ENABLE_PLAYGROUND` | `false` | Set to `true` to expose GraphQL Playground at `/` (off by default, it is not served otherwise) |
| `GRAPHQL_COMPLEXITY_LIMIT` | `5000` | Maximum query cost; list fields cost `limit x children`, so recursive fan-out is rejected |
| `NOTIFICATION_POLL_INTERVAL` | `3s` | How often a `notificationReceived` subscription polls the Notification service |

The gateway forwards a supplied bearer token to gRPC services as `authorization` metadata. It only trusts the user ID after verifying the JWT itself with `JWT_SECRET`.

## Build, generation, and tests

Run these commands from this directory in the complete Ouroboros workspace:

```sh
make generate
go test ./...
go run ./cmd
```

`make generate` refreshes `graph/generated/generated.go` and `graph/model/models_gen.go` from `graph/schema/*.graphqls`. Run code generation whenever the schema changes.

## Project layout

```
cmd/                 main.go: load config, dial backends, start the server
graph/               everything gqlgen owns, and nothing else
  schema/            *.graphqls, one file per domain (the API contract)
  generated/         generated executable schema (do not edit)
  model/             generated GraphQL models (do not edit)
  *.resolvers.go     thin resolvers: validate, authorize, delegate
  resolver.go        dependency container
internal/
  config/            environment -> Config
  clients/           typed gRPC clients for the six backend services
  auth/              JWT verification, HTTP + WebSocket auth, identity in context
  loaders/           per-operation data loaders + user assembly
  dataloader/        generic batching loader (no dependencies)
  notifications/     notifications read + polling subscription stream
  catalog/           optional Post/Recommendation capabilities (see below)
  mapping/           protobuf -> GraphQL model conversion
  args/              limit / ID-list validation
  server/            HTTP assembly, CORS, GraphQL handler, complexity limit
  testutil/          in-memory backend fakes for tests
```

Dependencies point one way: `cmd` -> `server` -> `graph` -> `loaders` /
`notifications` / `catalog` -> `clients` / `mapping` -> `auth` / `args`. Nothing
under `internal/` imports `graph` except `server`; `graph/model` is a leaf.

## Data loaders

Every GraphQL operation gets its own set of batching loaders (`internal/loaders`,
installed by `loaders.Extension`; generic implementation in `internal/dataloader`).
Nested fields (`Post.author`, `Comment.author`, `Notification.actor`,
`FeedItem.post`, `User.posts`, `Post.comments`) are lazy field resolvers, so all
keys requested at one level of a query are fetched with one downstream call,
de-duplicated and cached for the rest of that operation.

The previous code already batched the auth/profile lookups inside individual
resolvers, but could not share work *across* fields (e.g. post authors and
comment authors, or `User.posts` for many users). Now e.g. `postsByIds(20) { author
comments { author } }` costs one user lookup per nesting level, however many
posts/comments there are.

Per-user follow counts have no batch RPC, so they run concurrently (max 16 at a
time) instead of one after another. Subscriptions use non-caching loaders so a
long-lived stream never serves stale users. Mutations evict the affected user.

## Subscriptions

`notificationReceived(userId)` works over WebSocket (`graphql-ws`/`graphql-transport-ws`).
Browsers cannot set headers on a WebSocket, so send the token in
`connection_init`: `{"Authorization": "Bearer <jwt>"}`. The Notification service has
no streaming RPC, so the gateway polls it and emits each new notification once,
oldest first. Existing notifications are not replayed (use `notifications`). The
stream ends on disconnect or when the token's `exp` passes. The WebSocket origin
check uses `CORS_ALLOWED_ORIGINS`.

## Build, generation, and tests

Run these commands from this directory in the complete Ouroboros workspace:

```sh
go run github.com/99designs/gqlgen generate   # or `make generate`
go test -race ./...
go run ./cmd
```

The schema is split per domain (`graph/schema/*.graphqls`) so gqlgen's
follow-schema layout keeps one `<domain>.resolvers.go` per file. **Keep helpers,
constants and types out of `*.resolvers.go`**: gqlgen comments out anything in
those files that is not a resolver method. Put them in `resolver.go` or, better,
the matching `internal/` package.

## Fields that need new backend RPCs

`Query.posts`, `User.posts`, `Post.comments` and `Query.recommendations` are fully
implemented in the gateway, but the Post and Connection services do not expose the
list RPCs they need. By default `Query.posts` / `Query.recommendations` return an
explicit "unavailable" error and the nested lists are empty. See
[docs/proto-extensions.md](docs/proto-extensions.md) for the proposed RPCs and how
to enable them with `-tags postext` (adapter: `internal/catalog/grpc.go`).

## Verification status

The gateway was source-checked with `gofmt`. A complete build/test could not be run in this environment: the declared Go 1.25.8 toolchain was not available locally and toolchain downloads are blocked, and this archive omits the imported generated protobuf packages.




## Playground Examples
```
mutation signIn($input: SignInInput!) {
  signIn(input: $input) {
    token
    user {
      id
      email
      username
      displayName
    }
  }
}


{
  "input": {
    "email": "user1@example.com",
    "password": "user1user1"
  }
}
```

