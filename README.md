# ouroboros

A modern social networking platform built with a Next.js frontend and Go microservices, connected through a GraphQL API gateway and gRPC service-to-service communication.

The application supports core social workflows including authentication, profile management, posting, feed retrieval, recommendations, connect requests, and notifications. The GraphQL operations supplied for this draft are:

- GetFeed
- GetProfile
- UpdateProfile
- CreatePost
- LikePost
- GetUnreadNotifications
- GetRecommendations
- SendConnect
- SignUp
- SignIn
- SignOut

### What this documentation covers

- Project overview and architecture
- Local setup and deployment flow
- GraphQL operation reference
- gRPC service reference placeholder
- Frontend route reference placeholder
- Step-by-step tutorials
- Troubleshooting and FAQ
- Changelog structure by release

- **Go:** v1.21 or higher
- **Docker & Docker Compose:** Latest stable

### Installation

- Clone the repository.
- Install frontend dependencies.
- Install any service dependencies for the Go microservices.
- Configure environment variables.
- Start the frontend and backend services.

The frontend will be available at <http://localhost:3000>, and the GraphQL playground at <http://localhost:4000/graphql>.

### Environment variables

Create a .env.local file for the frontend and a matching environment file for services. The exact variable names should come from the repository, but the following categories are typically required:

- GraphQL gateway URL
- gRPC service hostnames and ports
- Authentication secrets
- Database connection strings
- Notification or message broker endpoints, if used
- Local development

Run the frontend and backend services in development mode.

### Frontend

npm install
npm run dev

### Example Go service

```
# Frontend
npm install
npm run dev

# Example Go service
go test ./...
go run ./cmd/<service-name>
```

### Deployment

Deployment should follow the repository's CI/CD process once provided. A standard flow is:

```
1 Build the frontend.
2. Build each Go service.
3. Run tests.
4. Run linting.
5. Publish container images.
6. Deploy to the target environment.
7. Verify GraphQL health and service readiness.
```

### Recommended verification steps

- Confirm the GraphQL gateway responds in staging.
- Confirm gRPC service health checks pass.
- Confirm the frontend can authenticate, fetch the feed, and create a post.
- Confirm notifications and profile queries return expected shapes.

# Reference

### GraphQL API

The GraphQL operations below are based only on the query and mutation definitions you supplied.

### Common patterns

IDs are passed as ID!.
Inputs are passed as GraphQL input objects where defined.
Mutations return only the fields selected in the operation.
Query results should be treated as the response contract for the frontend.

**_ link to GraphQL_API.md for the full reference of all operations _**

### GraphQL usage example in JavaScript

```

import { GraphQLClient } from "graphql-request"

```

### gRPC API

**_ link to gRPC_API.md for the full reference of all operations _**

### GraphQL usage example in Go

```

package main

```

# Troubleshooting & FAQ

After making changes to the GraphQL schema or gRPC proto files, ensure that all dependent services and frontend components are updated to match the new contract.

Commands in the makefile should be used to regenerate code and run tests after schema changes.

# Changelog / Release Notes

I actually don't know how to write these
