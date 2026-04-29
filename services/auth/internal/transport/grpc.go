package transport

import (
	"log"

	pb "ouroboros/proto/generated/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"context"
	"time"

	"google.golang.org/grpc/peer"
)

func StartGRPCServer(addr string, authServer pb.AuthServiceServer) {
	srv := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor),
	)

	reflection.Register(srv)

	pb.RegisterAuthServiceServer(srv, authServer)
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
