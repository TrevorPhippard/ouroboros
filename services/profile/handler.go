package main

import (
	"context"
	"fmt"
	"log"
	pb "ouroboros/proto/generated/profile"
)


type profileServiceServer struct {
	pb.UnimplementedProfileServiceServer
}

// GetProfile fetches a single user's profile information
func (s *profileServiceServer) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.Profile, error) {
	log.Printf("Profile Service: Fetching profile for User: %s", req.UserId)

	return &pb.Profile{
		UserId:      req.UserId,
		DisplayName: fmt.Sprintf("User %s", req.UserId),
		AvatarUrl:   fmt.Sprintf("https://api.dicebear.com/7.x/avataaars/svg?seed=%s", req.UserId),
		Bio:         "This is a mock bio for the Ouroboros social network.",
	}, nil
}

// GetProfilesByUserIds handles batch profile lookups
func (s *profileServiceServer) GetProfilesByUserIds(ctx context.Context, req *pb.GetProfilesByUserIdsRequest) (*pb.GetProfilesByUserIdsResponse, error) {
	log.Printf("Profile Service: Batch fetching %d profiles", len(req.UserIds))

	var profiles []*pb.Profile
	for _, id := range req.UserIds {
		profiles = append(profiles, &pb.Profile{
			UserId:      id,
			DisplayName: fmt.Sprintf("User %s", id),
			AvatarUrl:   fmt.Sprintf("https://api.dicebear.com/7.x/avataaars/svg?seed=%s", id),
			Bio:         "I am a user in the Ouroboros microservices ecosystem.",
		})
	}

	return &pb.GetProfilesByUserIdsResponse{
		Profiles: profiles,
	}, nil
}