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

// CRUD methods retained for compatibility with existing callers.

func (s *Service) CreateEntry(
	ctx context.Context,
	entry *domain.UserRecord,
) (*domain.UserRecord, error) {
	if entry == nil {
		return nil, errors.New("entry is required")
	}

	if entry.ID == "" {
		entry.ID = uuid.NewString()
	}

	if err := s.DB.WithContext(ctx).Create(entry).Error; err != nil {
		return nil, err
	}

	return entry, nil
}

func (s *Service) GetEntry(
	ctx context.Context,
	id string,
) (*domain.UserRecord, error) {
	var entry domain.UserRecord
	if err := s.DB.WithContext(ctx).
		First(&entry, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &entry, nil
}

func (s *Service) GetEntries(
	ctx context.Context,
	limit, offset int,
) ([]*domain.UserRecord, error) {
	if limit <= 0 {
		limit = 50
	}
	if offset < 0 {
		offset = 0
	}

	var entries []*domain.UserRecord
	if err := s.DB.WithContext(ctx).
		Limit(limit).
		Offset(offset).
		Find(&entries).Error; err != nil {
		return nil, err
	}

	return entries, nil
}

func (s *Service) UpdateEntry(
	ctx context.Context,
	entry *domain.UserRecord,
) (*domain.UserRecord, error) {
	if entry == nil {
		return nil, errors.New("entry is required")
	}

	result := s.DB.WithContext(ctx).
		Model(&domain.UserRecord{}).
		Where("id = ?", entry.ID).
		Update("display_name", entry.DisplayName)

	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	return s.GetEntry(ctx, entry.ID)
}

func (s *Service) DeleteEntry(
	ctx context.Context,
	id string,
) (*domain.UserRecord, error) {
	entry, err := s.GetEntry(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := s.DB.WithContext(ctx).Delete(entry).Error; err != nil {
		return nil, err
	}

	return entry, nil
}
