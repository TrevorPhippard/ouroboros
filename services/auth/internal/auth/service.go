package auth

import (
	"context"
	"encoding/json"
	"log"
	"time"

	pb "ouroboros/proto/generated/auth"

	"github.com/segmentio/kafka-go"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type Service struct {
	pb.UnimplementedAuthServiceServer
	Writer    *kafka.Writer
	DB        *gorm.DB
	JWTSecret []byte
}

func NewService(db *gorm.DB, writer *kafka.Writer, jwtSecret []byte) *Service {
	return &Service{
		DB:        db,
		Writer:    writer,
		JWTSecret: jwtSecret,
	}
}

func (s *Service) SignUp(ctx context.Context, req *pb.SignUpRequest) (*pb.AuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password required")
	}

	var existing UserRecord
	if err := s.DB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		return nil, status.Error(codes.AlreadyExists, "user already exists")
	}

	hash, err := HashPassword(req.Password)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to hash password")
	}

	user := UserRecord{
		ID:           generateID(),
		Email:        req.Email,
		DisplayName:  req.DisplayName,
		PasswordHash: hash,
		CreatedAt:    time.Now(),
	}

	if err := s.DB.Create(&user).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	token, err := GenerateToken(user.ID, s.JWTSecret)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	go s.publishUserSignedUp(context.Background(), &pb.User{
		Id:          user.ID,
		Email:       user.Email,
		DisplayName: user.DisplayName,
	})

	return &pb.AuthResponse{
		Token: token,
		User: &pb.User{
			Id:          user.ID,
			Email:       user.Email,
			DisplayName: user.DisplayName,
		},
	}, nil
}

func (s *Service) SignIn(ctx context.Context, req *pb.SignInRequest) (*pb.AuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password required")
	}

	var user UserRecord
	if err := s.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return nil, status.Error(codes.NotFound, "invalid credentials")
	}

	if err := CheckPassword(req.Password, user.PasswordHash); err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	token, err := GenerateToken(user.ID, s.JWTSecret)
	if err != nil {
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

func (s *Service) SignOut(ctx context.Context, req *pb.SignOutRequest) (*pb.SignOutResponse, error) {
	log.Println("SignOut called")
	return &pb.SignOutResponse{Success: true}, nil
}

func (s *Service) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	return nil, status.Error(codes.Unimplemented, "get user is not wired to a persistent user store yet")
}

func (s *Service) GetUsersByIds(ctx context.Context, req *pb.GetUsersByIdsRequest) (*pb.GetUsersByIdsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "get users by ids is not wired to a persistent user store yet")
}

func (s *Service) publishUserSignedUp(ctx context.Context, user *pb.User) error {
	if s.Writer == nil || user == nil {
		return nil
	}

	event := newUserSignedUpEvent(user.Id, user.Email, user.DisplayName)

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return s.Writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(user.Id),
		Value: payload,
	})
}
