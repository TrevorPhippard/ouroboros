package graph

import (
	pb "ouroboros/proto" // Adjust this to match your actual proto module path
)

// This file will not be regenerated automatically.
//
// It serves as dependency injection for your app, add any dependencies you require here.

type Resolver struct {
	UserServiceClient pb.Service1Client
	TodoServiceClient pb.Service2Client
}