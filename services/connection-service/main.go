package main

import (
	"context"
	"fmt"
	"log"
	"net"

	pb "ouroboros/proto"

	"google.golang.org/grpc"
)

type connServiceServer struct {
	pb.UnimplementedService2Server
}

func (s *connServiceServer) BatchGetTest2(ctx context.Context, req *pb.BatchRequest) (*pb.Test2BatchResponse, error) {
	log.Printf("Connection Service: Fetching %d IDs", len(req.Ids))

	var items []*pb.Test2
	for _, id := range req.Ids {
		items = append(items, &pb.Test2{
			Id:         id,
			MockData_2: fmt.Sprintf("Connection Task Data for %s", id),
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
	pb.RegisterService2Server(s, &connServiceServer{})

	log.Println("Connection Service (gRPC) running on :50053")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}