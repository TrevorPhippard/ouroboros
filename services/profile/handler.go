package main

import (
	"context"
	"log"
	pb "ouroboros/proto/generated/profile"

	"gorm.io/gorm"
)


type profileServiceServer struct {
	pb.UnimplementedProfileServiceServer
	db *gorm.DB
}

// GetProfile fetches a single user's profile information
func (s *profileServiceServer) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.Profile, error) {
	log.Printf("Profile Service: Fetching profile for User: %s", req.UserId)

	var profile Profile
	if err := s.db.Where("user_id = ?", req.UserId).First(&profile).Error; err != nil {
		return nil, err
	}

	return &pb.Profile{
		Id:          profile.ID,
		UserId:      profile.UserId,
		DisplayName: profile.DisplayName,
		AvatarUrl:   profile.AvatarUrl,
		Bio:         profile.Bio,
		Headline:    profile.Headline,
		About:       profile.About,
		// Experiences: convert to pb.Experience
	}, nil
}

// GetProfilesByUserIds handles batch profile lookups
func (s *profileServiceServer) GetProfilesByUserIds(ctx context.Context, req *pb.GetProfilesByUserIdsRequest) (*pb.GetProfilesByUserIdsResponse, error) {
	log.Printf("Profile Service: Batch fetching %d profiles", len(req.UserIds))

	var profiles []Profile
	if err := s.db.Where("user_id IN ?", req.UserIds).Find(&profiles).Error; err != nil {
		return nil, err
	}

	var pbProfiles []*pb.Profile
	for _, p := range profiles {
		pbProfiles = append(pbProfiles, &pb.Profile{
			Id:          p.ID,
			UserId:      p.UserId,
			DisplayName: p.DisplayName,
			AvatarUrl:   p.AvatarUrl,
			Bio:         p.Bio,
			Headline:    p.Headline,
			About:       p.About,
			// Experiences
		})
	}

	return &pb.GetProfilesByUserIdsResponse{
		Profiles: pbProfiles,
	}, nil
}

// UpdateProfile updates a user's profile
func (s *profileServiceServer) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.UpdateProfileResponse, error) {
	log.Printf("Profile Service: Updating profile for User: %s", req.UserId)

	var profile Profile
	if err := s.db.Where("user_id = ?", req.UserId).First(&profile).Error; err != nil {
		return nil, err
	}

	// Update fields
	if req.Headline != "" {
		profile.Headline = req.Headline
	}
	if req.About != "" {
		profile.About = req.About
	}

	if err := s.db.Save(&profile).Error; err != nil {
		return nil, err
	}

	return &pb.UpdateProfileResponse{
		Profile: &pb.Profile{
			Id:          profile.ID,
			UserId:      profile.UserId,
			DisplayName: profile.DisplayName,
			AvatarUrl:   profile.AvatarUrl,
			Bio:         profile.Bio,
			Headline:    profile.Headline,
			About:       profile.About,
		},
	}, nil
}