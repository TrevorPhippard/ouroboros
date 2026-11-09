package transport

import (
	"context"
	"errors"
	"log"
	"net"
	"time"

	pb "ouroboros/proto/generated/auth"

	"google.golang.org/grpc"
	"google.golang.org/grpc/peer"
	"google.golang.org/grpc/reflection"
)

func StartGRPCServer(
	ctx context.Context,
	addr string,
	authServer pb.AuthServiceServer,
) error {
	lis, err := net.Listen("tcp", addr)
	if err != nil {
		log.Printf("gRPC failed to listen on %s: %v", addr, err)
		return err
	}

	log.Printf("gRPC server listening on %s", addr)

	srv := grpc.NewServer(
		grpc.UnaryInterceptor(loggingInterceptor),
	)

	reflection.Register(srv)
	pb.RegisterAuthServiceServer(srv, authServer)

	go func() {
		<-ctx.Done()
		srv.GracefulStop()
	}()

	log.Printf("gRPC server listening on %s", addr)

	err = srv.Serve(lis)
	if errors.Is(err, grpc.ErrServerStopped) {
		return nil
	}

	return err
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
