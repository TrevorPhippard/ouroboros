package main

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"connection/internal/consul"
	pb "ouroboros/proto/generated/connection"

	"github.com/hashicorp/consul/api"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

const connectionDataPath = "/tmp/ouroboros_connection_store.json"

type connectionStore struct {
	mu        sync.RWMutex
	path      string
	Followers map[string]map[string]bool `json:"followers"`
}

type connServiceServer struct {
	pb.UnimplementedConnectionServiceServer
	store *connectionStore
}

func newConnectionStore(path string) (*connectionStore, error) {
	store := &connectionStore{
		path:      path,
		Followers: map[string]map[string]bool{},
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}

	if err := store.load(); err != nil {
		return nil, err
	}
	if len(store.Followers) == 0 {
		store.seed()
		if err := store.persistLocked(); err != nil {
			return nil, err
		}
	}

	return store, nil
}

func (s *connectionStore) load() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	data, err := os.ReadFile(s.path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if len(data) == 0 {
		return nil
	}

	return json.Unmarshal(data, s)
}

func (s *connectionStore) seed() {
	s.Followers["user-1"] = map[string]bool{"user-2": true, "user-3": true}
	s.Followers["user-2"] = map[string]bool{"user-1": true}
	s.Followers["user-3"] = map[string]bool{"user-1": true}
}

func (s *connectionStore) persistLocked() error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

func (s *connectionStore) follow(followerID, followeeID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.Followers[followeeID] == nil {
		s.Followers[followeeID] = map[string]bool{}
	}
	s.Followers[followeeID][followerID] = true
	return s.persistLocked()
}

func (s *connectionStore) unfollow(followerID, followeeID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if followers := s.Followers[followeeID]; followers != nil {
		delete(followers, followerID)
	}
	return s.persistLocked()
}

func (s *connectionStore) followersCount(userID string) int32 {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return int32(len(s.Followers[userID]))
}

func (s *connectionStore) followingCount(userID string) int32 {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var count int32
	for _, followers := range s.Followers {
		if followers[userID] {
			count++
		}
	}
	return count
}

func (s *connectionStore) isFollowing(followerID, followeeID string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.Followers[followeeID] != nil && s.Followers[followeeID][followerID]
}

func (s *connectionStore) followeesForUser(userID string) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	followees := make([]string, 0)
	for followeeID, followers := range s.Followers {
		if followers[userID] {
			followees = append(followees, followeeID)
		}
	}
	sort.Strings(followees)
	return followees
}

func (s *connServiceServer) FollowUser(ctx context.Context, req *pb.FollowUserRequest) (*pb.FollowUserResponse, error) {
	if req == nil || strings.TrimSpace(req.FollowerId) == "" || strings.TrimSpace(req.FolloweeId) == "" {
		return nil, status.Error(codes.InvalidArgument, "follower_id and followee_id are required")
	}
	if req.FollowerId == req.FolloweeId {
		return nil, status.Error(codes.InvalidArgument, "cannot follow yourself")
	}

	if err := s.store.follow(req.FollowerId, req.FolloweeId); err != nil {
		log.Printf("connection-service: failed to persist follow follower=%s followee=%s: %v", req.FollowerId, req.FolloweeId, err)
		return nil, status.Error(codes.Internal, "failed to create connection")
	}

	log.Printf("connection-service: %s followed %s", req.FollowerId, req.FolloweeId)
	return &pb.FollowUserResponse{Success: true}, nil
}

func (s *connServiceServer) UnfollowUser(ctx context.Context, req *pb.UnfollowUserRequest) (*pb.UnfollowUserResponse, error) {
	if req == nil || strings.TrimSpace(req.FollowerId) == "" || strings.TrimSpace(req.FolloweeId) == "" {
		return nil, status.Error(codes.InvalidArgument, "follower_id and followee_id are required")
	}

	if err := s.store.unfollow(req.FollowerId, req.FolloweeId); err != nil {
		log.Printf("connection-service: failed to persist unfollow follower=%s followee=%s: %v", req.FollowerId, req.FolloweeId, err)
		return nil, status.Error(codes.Internal, "failed to remove connection")
	}

	log.Printf("connection-service: %s unfollowed %s", req.FollowerId, req.FolloweeId)
	return &pb.UnfollowUserResponse{Success: true}, nil
}

func (s *connServiceServer) GetFollowersCount(ctx context.Context, req *pb.UserRequest) (*pb.CountResponse, error) {
	if req == nil || strings.TrimSpace(req.UserId) == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	return &pb.CountResponse{Count: s.store.followersCount(req.UserId)}, nil
}

func (s *connServiceServer) GetFollowingCount(ctx context.Context, req *pb.UserRequest) (*pb.CountResponse, error) {
	if req == nil || strings.TrimSpace(req.UserId) == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}
	return &pb.CountResponse{Count: s.store.followingCount(req.UserId)}, nil
}

func (s *connServiceServer) IsFollowing(ctx context.Context, req *pb.IsFollowingRequest) (*pb.IsFollowingResponse, error) {
	if req == nil || strings.TrimSpace(req.FollowerId) == "" || strings.TrimSpace(req.FolloweeId) == "" {
		return nil, status.Error(codes.InvalidArgument, "follower_id and followee_id are required")
	}
	return &pb.IsFollowingResponse{IsFollowing: s.store.isFollowing(req.FollowerId, req.FolloweeId)}, nil
}

func main() {
	store, err := newConnectionStore(connectionDataPath)
	if err != nil {
		log.Fatalf("failed to initialize connection store: %v", err)
	}

	addr := "consul:8500"

	agent := consul.NewAgent(&api.Config{
		Address: addr,
	})

	serviceCfg := consul.Config{
		ServiceID:   "connection-service-1",
		ServiceName: "connection-service",
		Address:     "connection-service",
		Tags:        []string{"grpc", "connection"},
		Port:        50054,
		Check: &api.AgentServiceCheck{
			HTTP:     "http://connection-service:8080/health",
			Interval: "10s",
			Timeout:  "2s",
		},
	}
	if err := agent.RegisterService(serviceCfg); err != nil {
		log.Fatalf("failed to register service: %v", err)
	}

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

	lis, err := net.Listen("tcp", ":50054")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterConnectionServiceServer(s, &connServiceServer{store: store})
	reflection.Register(s)

	log.Println("Connection Service (gRPC) running on :50054")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
