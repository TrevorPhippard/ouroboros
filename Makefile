.PHONY: infra-up infra-down proto tidy api-gateway auth-service

# Starts just the databases and Kafka
infra-up:
	docker-compose up -d postgres neo4j redis zookeeper kafka

# Tears down infra and wipes volumes (careful!)
infra-down:
	docker-compose down -v

# Generates Go code from gRPC proto files
proto:
	@echo "Generating gRPC code..."
		mkdir -p proto/generated
		protoc --proto_path=proto \
				--go_out=proto/generated --go_opt=paths=source_relative \
				--go-grpc_out=proto/generated --go-grpc_opt=paths=source_relative \
				$(shell find proto -name "*.proto")


# Generate graphql schema
graphql:
	@echo "Generating GraphQL schema..."
	go run github.com/99designs/gqlgen generate

# Syncs Go dependencies across the workspace
tidy:
	@echo "Tidying workspace..."
	for d in services/* pkg; do \
		(cd $$d && go mod tidy); \
	done

# --- Run Services Locally (Bare Metal) ---
api-gateway:
	go run services/api-gateway/cmd/server/main.go

auth-service:
	go run services/auth/cmd/server/main.go