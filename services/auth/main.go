package main

import (
	"context"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"auth/internal/database"
	"auth/internal/discovery"
	"auth/internal/messaging"
	pb "ouroboros/proto/generated/auth"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/segmentio/kafka-go"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
	"gorm.io/gorm"
)

type authServiceServer struct {
	pb.UnimplementedAuthServiceServer
	Writer *kafka.Writer
	DB     *gorm.DB
}

type userSignedUpEvent struct {
	EventID   string `json:"eventId"`
	Type      string `json:"type"`
	Timestamp string `json:"timestamp"`
	Data      struct {
		UserID      string `json:"userId"`
		Email       string `json:"email"`
		DisplayName string `json:"displayName"`
	} `json:"data"`
}

var jwtSecret = []byte("super-secret") // move to env in prod

type userRecord struct {
	ID           string
	Email        string
	DisplayName  string
	PasswordHash string
	CreatedAt    int64
}

func (userRecord) TableName() string {
	return "users"
}

func generateID() string {
	return uuid.NewString()
}

func HashPassword(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	return string(bytes), err
}

func CheckPassword(password, hash string) error {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
}

func GenerateToken(userID string) (string, error) {
	claims := jwt.MapClaims{
		"userId": userID,
		"exp":    time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func (s *authServiceServer) SignUp(ctx context.Context, req *pb.SignUpRequest) (*pb.AuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password required")
	}

	// Check if user exists
	var existing userRecord
	if err := s.DB.Where("email = ?", req.Email).First(&existing).Error; err == nil {
		return nil, status.Error(codes.AlreadyExists, "user already exists")
	}

	// Hash password
	hash, err := HashPassword(req.Password)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to hash password")
	}

	user := userRecord{
		ID:           generateID(), // use uuid
		Email:        req.Email,
		DisplayName:  req.DisplayName,
		PasswordHash: hash,
		CreatedAt:    time.Now().Unix(),
	}

	if err := s.DB.Create(&user).Error; err != nil {
		return nil, status.Error(codes.Internal, "failed to create user")
	}

	// Generate token
	token, err := GenerateToken(user.ID)
	if err != nil {
		return nil, status.Error(codes.Internal, "failed to generate token")
	}

	// Publish event (async side-effect)
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

func (s *authServiceServer) SignIn(ctx context.Context, req *pb.SignInRequest) (*pb.AuthResponse, error) {
	if req.Email == "" || req.Password == "" {
		return nil, status.Error(codes.InvalidArgument, "email and password required")
	}

	var user userRecord
	if err := s.DB.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return nil, status.Error(codes.NotFound, "invalid credentials")
	}

	// Compare password
	if err := CheckPassword(req.Password, user.PasswordHash); err != nil {
		return nil, status.Error(codes.Unauthenticated, "invalid credentials")
	}

	// Generate token
	token, err := GenerateToken(user.ID)
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

func (s *authServiceServer) publishUserSignedUp(ctx context.Context, user *pb.User) error {
	if s.Writer == nil || user == nil {
		return nil
	}

	event := userSignedUpEvent{
		EventID:   "signup-" + user.Id,
		Type:      "UserSignedUp",
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}
	event.Data.UserID = user.Id
	event.Data.Email = user.Email
	event.Data.DisplayName = user.DisplayName

	payload, err := json.Marshal(event)
	if err != nil {
		return err
	}

	return s.Writer.WriteMessages(ctx, kafka.Message{
		Key:   []byte(user.Id),
		Value: payload,
	})
}

func (s *authServiceServer) SignOut(ctx context.Context, req *pb.SignOutRequest) (*pb.SignOutResponse, error) {
	log.Println("SignOut called")
	return &pb.SignOutResponse{
		Success: true,
	}, nil
}

func (s *authServiceServer) GetUser(ctx context.Context, req *pb.GetUserRequest) (*pb.User, error) {
	return nil, status.Error(codes.Unimplemented, "get user is not wired to a persistent user store yet")
}

func (s *authServiceServer) GetUsersByIds(ctx context.Context, req *pb.GetUsersByIdsRequest) (*pb.GetUsersByIdsResponse, error) {
	return nil, status.Error(codes.Unimplemented, "get users by ids is not wired to a persistent user store yet")
}

func main() {
	// 1. Setup Infrastructure
	dbURL := os.Getenv("DB_URL")
	db := database.Connect(dbURL)
	database.Migrate(db)
	database.SeedDB(db)

	discovery.Register("consul:8500", "auth-service", 50053)

	writer := messaging.NewPostProducer("kafka:9092", "user.signed_up")
	defer writer.Close()

	// 2. Start HTTP Server (Metrics + Health)
	go func() {
		http.Handle("/metrics", promhttp.Handler())

		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte("OK"))
		})

		log.Println("HTTP server running on :8080 (metrics + health)")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("failed to start HTTP server: %v", err)
		}
	}()

	// 3. Start gRPC Server

	lis, err := net.Listen("tcp", ":50053")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterAuthServiceServer(s, &authServiceServer{Writer: writer, DB: db})
	reflection.Register(s)
	log.Println("Auth Service (gRPC) running on :50053")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
