# ==========================================
# Stage 1: The Builder
# ==========================================
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Disable Go Workspace mode for the build process
# This prevents go.work from interfering with individual service builds
ENV GOWORK=off

COPY go.mod go.sum ./
RUN go mod download

COPY . .

ARG SERVICE_DIR

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux go build -o /app/bin/service ./services/${SERVICE_DIR}/main.go

# ==========================================
# Stage 2: The Production Image (Alpine)
# ==========================================
FROM alpine:3.19
RUN apk --no-cache add ca-certificates
WORKDIR /app
COPY --from=builder /app/bin/service .

CMD ["./service"]