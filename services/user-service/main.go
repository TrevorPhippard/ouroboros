package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "ouroboros/proto"

	"google.golang.org/grpc"
)

type userServiceServer struct {
	pb.UnimplementedService1Server
}

func (s *userServiceServer) BatchGetTest1(ctx context.Context, req *pb.BatchRequest) (*pb.Test1BatchResponse, error) {
	log.Printf("User Service: Fetching %d IDs", len(req.Ids))

	var items []*pb.Test1
	for _, id := range req.Ids {
		items = append(items, &pb.Test1{
			Id:         id,
			MockData_1: fmt.Sprintf("User Profile Data for %s", id),
		})
	}

	return &pb.Test1BatchResponse{Items: items}, nil
}

func main() {
	lis, err := net.Listen("tcp", ":50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterService1Server(s, &userServiceServer{})

	log.Println("User Service (gRPC) running on :50051")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}