# ==========================================
# Stage 1: The Builder
# ==========================================
FROM golang:1.25-alpine AS builder

WORKDIR /app

# 1. Copy workspace and ALL go.mod files
# We use a wildcard to grab any go.mod in any subfolder under services/
COPY go.work ./
COPY go.mod* go.sum* ./
COPY services/api-gateway/go.mod* services/api-gateway/go.sum* ./services/api-gateway/
COPY services/auth/go.mod* services/auth/go.sum* ./services/auth/
COPY services/connection/go.mod* services/connection/go.sum* ./services/connection/
COPY services/feed/go.mod* services/feed/go.sum* ./services/feed/
COPY services/notification/go.mod* services/notification/go.sum* ./services/notification/
COPY services/post/go.mod* services/post/go.sum* ./services/post/
COPY services/profile/go.mod* services/profile/go.sum* ./services/profile/

# Download dependencies for the entire workspace
RUN go mod download

# 2. Copy the actual source code
COPY . .

# 3. Build the specific service
ARG SERVICE_DIR
# We use the full path to the main.go based on your structure
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/service ./services/${SERVICE_DIR}

# ==========================================
# Stage 2: The Production Image
# ==========================================
FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/bin/service .

CMD ["./service"]