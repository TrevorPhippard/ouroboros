package service

import (
	"context"
	"errors"
	"log"
	"strconv"
	"strings"
	"time"

	pb "ouroboros/proto/generated/profile"
	"profile/internal/models"

	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type ProfileServiceServer struct {
	pb.UnimplementedProfileServiceServer
	DB     *gorm.DB
	Writer *kafka.Writer
}

func (s *ProfileServiceServer) GetProfile(ctx context.Context, req *pb.GetProfileRequest) (*pb.Profile, error) {
	if req == nil || strings.TrimSpace(req.UserId) == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	var profile models.Profile
	if err := s.DB.WithContext(ctx).Preload("Experiences").First(&profile, "user_id = ?", req.UserId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "profile not found")
		}
		log.Printf("profile-service: failed to fetch profile user_id=%s: %v", req.UserId, err)
		return nil, status.Error(codes.Internal, "failed to fetch profile")
	}

	return toProtoProfile(&profile), nil
}

func (s *ProfileServiceServer) GetProfilesByUserIds(ctx context.Context, req *pb.GetProfilesByUserIdsRequest) (*pb.GetProfilesByUserIdsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "request is required")
	}

	userIDs := uniqueNonEmpty(req.UserIds)
	if len(userIDs) == 0 {
		return &pb.GetProfilesByUserIdsResponse{Profiles: []*pb.Profile{}}, nil
	}

	var profiles []models.Profile
	if err := s.DB.WithContext(ctx).Preload("Experiences").Where("user_id IN ?", userIDs).Find(&profiles).Error; err != nil {
		log.Printf("profile-service: failed to batch fetch %d profiles: %v", len(userIDs), err)
		return nil, status.Error(codes.Internal, "failed to fetch profiles")
	}

	byUserID := make(map[string]*models.Profile, len(profiles))
	for i := range profiles {
		byUserID[profiles[i].UserId] = &profiles[i]
	}

	result := make([]*pb.Profile, 0, len(userIDs))
	for _, userID := range userIDs {
		if profile := byUserID[userID]; profile != nil {
			result = append(result, toProtoProfile(profile))
		}
	}

	return &pb.GetProfilesByUserIdsResponse{Profiles: result}, nil
}

func (s *ProfileServiceServer) UpdateProfile(ctx context.Context, req *pb.UpdateProfileRequest) (*pb.UpdateProfileResponse, error) {
	if req == nil || strings.TrimSpace(req.UserId) == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	var profile models.Profile
	if err := s.DB.WithContext(ctx).First(&profile, "user_id = ?", req.UserId).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.NotFound, "profile not found")
		}
		log.Printf("profile-service: failed to load profile for update user_id=%s: %v", req.UserId, err)
		return nil, status.Error(codes.Internal, "failed to load profile")
	}

	if headline := strings.TrimSpace(req.Headline); headline != "" {
		profile.Headline = headline
	}
	if about := strings.TrimSpace(req.About); about != "" {
		profile.About = about
	}

	if err := s.DB.WithContext(ctx).Save(&profile).Error; err != nil {
		log.Printf("profile-service: failed to update profile user_id=%s: %v", req.UserId, err)
		return nil, status.Error(codes.Internal, "failed to update profile")
	}

	return &pb.UpdateProfileResponse{Profile: toProtoProfile(&profile)}, nil
}

func toProtoProfile(profile *models.Profile) *pb.Profile {
	experiences := make([]*pb.Experience, 0, len(profile.Experiences))
	for i := range profile.Experiences {
		experience := &pb.Experience{
			Id:          strconv.FormatUint(uint64(profile.Experiences[i].ID), 10),
			Title:       profile.Experiences[i].Title,
			Company:     profile.Experiences[i].Company,
			StartDate:   profile.Experiences[i].StartDate.UTC().Format(time.RFC3339),
			Description: "",
		}
		if profile.Experiences[i].EndDate != nil {
			experience.EndDate = profile.Experiences[i].EndDate.UTC().Format(time.RFC3339)
		}
		experiences = append(experiences, experience)
	}

	return &pb.Profile{
		Id:          profile.ID,
		UserId:      profile.UserId,
		Headline:    profile.Headline,
		About:       profile.About,
		DisplayName: profile.DisplayName,
		AvatarUrl:   profile.AvatarUrl,
		Bio:         profile.Bio,
		Experiences: experiences,
	}
}

func uniqueNonEmpty(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	result := make([]string, 0, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		result = append(result, value)
	}
	return result
}
