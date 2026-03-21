package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "ouroboros/proto/generated/auth"

	"google.golang.org/grpc"
)

type authServiceServer struct {
	pb.UnimplementedAuthServiceServer
}

func (s *authServiceServer) BatchGetTest2(ctx context.Context, req *pb.BatchRequest) (*pb.Test2BatchResponse, error) {
	log.Printf("Auth Service: Fetching %d IDs", len(req.Ids))

	var items []*pb.Test2
	for _, id := range req.Ids {
		items = append(items, &pb.Test2{
			Id:         id,
			MockData_2: fmt.Sprintf("Auth Task Data for %s", id),
		})
	}

	return &pb.Test2BatchResponse{Items: items}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterService2Server(s, &authServiceServer{})

	log.Println("Auth Service (gRPC) running on :50053")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}