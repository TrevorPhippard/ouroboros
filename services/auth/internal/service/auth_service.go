package services

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"auth/internal/domain"
	messagingkafka "auth/internal/messaging/kafka"

	pb "ouroboros/proto/generated/auth"

	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type Service struct {
	pb.UnimplementedAuthServiceServer

	Writer    *messagingkafka.Producer
	DB        *gorm.DB
	JWTSecret []byte
	topic     string
	logger    *slog.Logger
}

func NewService(
	db *gorm.DB,
	writer *messagingkafka.Producer,
	jwtSecret []byte,
	topic string,
	logger *slog.Logger,
) *Service {

	return &Service{
		DB:        db,
		Writer:    writer,
		JWTSecret: jwtSecret,
		topic:     topic,
		logger:    logger,
	}
}

func (s *Service) SignUp(
	ctx context.Context,
	req *pb.SignUpRequest,
) (*pb.AuthResponse, error) {
	if req == nil || req.Email == "" || req.Password == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"email and password required",
		)
	}

	var existing domain.UserRecord
	err := s.DB.WithContext(ctx).
		Where("email = ?", req.Email).
		First(&existing).Error

	if err == nil {
		return nil, status.Error(codes.AlreadyExists, "user already exists")
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Error("Failed to check existing user", "error", err)
		return nil, status.Error(codes.Internal, "database error")
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		s.logger.Error("Failed to hash password", "error", err)
		return nil, status.Error(codes.Internal, "failed to hash password")
	}

	user := domain.UserRecord{
		ID:           uuid.NewString(),
		Email:        req.Email,
		DisplayName:  req.DisplayName,
		PasswordHash: hash,
		CreatedAt:    time.Now(),
	}

	if err := s.DB.WithContext(ctx).Create(&user).Error; err != nil {
		s.logger.Error("Failed to create user", "error", err)
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	token, err := GenerateToken(user.ID, s.JWTSecret)
	if err != nil {
		s.logger.Error("Failed to generate token", "error", err)
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	event := messagingkafka.UserCreatedEvent{
		ID:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		CreatedAt:   user.CreatedAt,
	}

	go func() {
		publishCtx, cancel := context.WithTimeout(
			context.Background(),
			5*time.Second,
		)
		defer cancel()

		if s.Writer == nil {
			s.logger.Warn("Kafka producer unavailable; signup event not published")
			return
		}

		if err := s.Writer.Publish(publishCtx, s.topic, event); err != nil {
			s.logger.Error("Failed to publish signup event", "error", err)
		}
	}()

	return &pb.AuthResponse{
		Token: token,
		User: &pb.User{
			Id:          user.ID,
			Email:       user.Email,
			DisplayName: user.DisplayName,
		},
	}, nil
}

func (s *Service) SignIn(
	ctx context.Context,
	req *pb.SignInRequest,
) (*pb.AuthResponse, error) {
	if req == nil || req.Email == "" || req.Password == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"email and password required",
		)
	}

	var user domain.UserRecord
	if err := s.DB.WithContext(ctx).
		Where("email = ?", req.Email).
		First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(codes.Unauthenticated, "invalid credentials")
		}
		s.logger.Error("Failed to find user", "error", err)
		return nil, status.Error(codes.Internal, "database error")
	}

	if err := CheckPassword(req.Password, user.PasswordHash); err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	token, err := GenerateToken(user.ID, s.JWTSecret)
	if err != nil {
		s.logger.Error("Failed to generate token", "error", err)
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	return &pb.AuthResponse{
		Token: token,
		User: &pb.User{
			Id:          user.ID,
			Email:       user.Email,
			DisplayName: user.DisplayName,
		},
	}, nil
}

func (s *Service) SignOut(
	ctx context.Context,
	req *pb.SignOutRequest,
) (*pb.SignOutResponse, error) {
	s.logger.Info("SignOut called")
	return &pb.SignOutResponse{Success: true}, nil
}

func (s *Service) GetUser(
	ctx context.Context,
	req *pb.GetUserRequest,
) (*pb.User, error) {
	if req == nil || req.Id == "" {
		return nil, status.Error(
			codes.InvalidArgument,
			"user id is required",
		)
	}

	var user domain.UserRecord
	err := s.DB.WithContext(ctx).
		Where("id = ?", req.Id).
		First(&user).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, status.Error(
				codes.NotFound,
				"user not found",
			)
		}

		s.logger.Error(
			"Failed to get user",
			"user_id", req.Id,
			"error", err,
		)
		return nil, status.Error(
			codes.Internal,
			"database error",
		)
	}

	return &pb.User{
		Id:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
	}, nil
}

func (s *Service) GetUsersByIds(
	ctx context.Context,
	req *pb.GetUsersByIdsRequest,
) (*pb.GetUsersByIdsResponse, error) {
	if req == nil || len(req.Ids) == 0 {
		return nil, status.Error(
			codes.InvalidArgument,
			"user ids are required",
		)
	}

	// Remove empty IDs and duplicates while preserving input order.
	ids := make([]string, 0, len(req.Ids))
	seen := make(map[string]struct{}, len(req.Ids))

	for _, id := range req.Ids {
		if id == "" {
			continue
		}

		if _, exists := seen[id]; exists {
			continue
		}

		seen[id] = struct{}{}
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return nil, status.Error(
			codes.InvalidArgument,
			"at least one valid user id is required",
		)
	}

	var records []domain.UserRecord
	err := s.DB.WithContext(ctx).
		Where("id IN ?", ids).
		Find(&records).Error

	if err != nil {
		s.logger.Error(
			"Failed to get users by ids",
			"error", err,
		)
		return nil, status.Error(
			codes.Internal,
			"database error",
		)
	}

	// Index results by ID because SQL IN queries do not guarantee order.
	usersByID := make(map[string]*pb.User, len(records))

	for _, record := range records {
		usersByID[record.ID] = &pb.User{
			Id:          record.ID,
			Email:       record.Email,
			DisplayName: record.DisplayName,
		}
	}

	// Return existing users in the same order as the requested IDs.
	users := make([]*pb.User, 0, len(records))
	for _, id := range ids {
		if user, exists := usersByID[id]; exists {
			users = append(users, user)
		}
	}

	return &pb.GetUsersByIdsResponse{
		Users: users,
	}, nil
}
