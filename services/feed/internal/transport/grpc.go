package transport

import (
	"context"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
)

func NewServer() *grpc.Server {
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor),
	)

	reflection.Register(srv)

	return srv
}

func loggingInterceptor(
	ctx context.Context,
	req any,
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (any, error) {
	start := time.Now()
	p, _ := peer.FromContext(ctx)

	log.Printf("[gRPC] %s from %v", info.FullMethod, p.Addr)

	resp, err := handler(ctx, req)

	log.Printf("[gRPC DONE] %s in %s", info.FullMethod, time.Since(start))

	return resp, err
}
