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
| `JWT_SECRET` | unset | Shared HS256 verification secret; protected operations fail closed if unset |
| `CORS_ALLOWED_ORIGINS` | `http://localhost:3000,http://127.0.0.1:3000` | Comma-separated browser origins |
| `ENABLE_PLAYGROUND` | `false` | Set to `true` to expose GraphQL Playground at `/` |

The gateway forwards a supplied bearer token to gRPC services as `authorization` metadata. It only trusts the user ID after verifying the JWT itself with `JWT_SECRET`.

## Build, generation, and tests

Run these commands from this directory in the complete Ouroboros workspace:

```sh
make generate
go test ./...
go run ./cmd
```

`make generate` refreshes `graph/generated.go` and `graph/model/models_gen.go` from `graph/schema.graphqls`. The generated files were inconsistent with the checked-in schema in the supplied archive, so code generation should be run whenever the schema changes.

## Known backend gaps

The gateway schema exposes `Query.posts`, but the available post client calls in this project only show `GetPost` and `GetPostsByIds`. The resolver therefore returns an explicit error until the Post service provides a list-posts RPC. The nested `User.posts` and `Post.comments` fields also need real list RPCs to return data; they currently default to empty lists. `recommendations` and notification subscriptions are not implemented by the current downstream service interfaces. Connect these fields to real backend RPCs before clients depend on them.

The gateway was source-checked with `gofmt`. A complete build/test could not be run in this environment: the declared Go 1.25.8 toolchain was not available locally and toolchain downloads are blocked, and this archive omits the imported generated protobuf packages.
