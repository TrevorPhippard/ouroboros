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

	"notification/internal/consul"
	pb "ouroboros/proto/generated/notification"

	"github.com/hashicorp/consul/api"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/reflection"
	"google.golang.org/grpc/status"
)

const notificationDataPath = "/tmp/ouroboros_notification_store.json"

type notificationRecord struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	Type      string `json:"type"`
	ActorID   string `json:"actor_id"`
	EntityID  string `json:"entity_id"`
	CreatedAt string `json:"created_at"`
	Read      bool   `json:"read"`
}

type notificationStore struct {
	mu            sync.RWMutex
	path          string
	Notifications []notificationRecord `json:"notifications"`
}

type notificationServiceServer struct {
	pb.UnimplementedNotificationServiceServer
	store *notificationStore
}

func newNotificationStore(path string) (*notificationStore, error) {
	store := &notificationStore{
		path:          path,
		Notifications: []notificationRecord{},
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	if err := store.load(); err != nil {
		return nil, err
	}
	if len(store.Notifications) == 0 {
		store.seed()
		if err := store.persistLocked(); err != nil {
			return nil, err
		}
	}
	return store, nil
}

func (s *notificationStore) load() error {
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

func (s *notificationStore) seed() {
	s.Notifications = []notificationRecord{
		{
			ID:        "notif-1",
			UserID:    "user-1",
			Type:      "FOLLOW",
			ActorID:   "user-2",
			EntityID:  "user-2",
			CreatedAt: "2026-04-27T10:00:00Z",
			Read:      false,
		},
		{
			ID:        "notif-2",
			UserID:    "user-1",
			Type:      "FOLLOW",
			ActorID:   "user-3",
			EntityID:  "user-3",
			CreatedAt: "2026-04-27T10:05:00Z",
			Read:      false,
		},
		{
			ID:        "notif-3",
			UserID:    "user-2",
			Type:      "FOLLOW",
			ActorID:   "user-1",
			EntityID:  "user-1",
			CreatedAt: "2026-04-27T10:10:00Z",
			Read:      true,
		},
	}
}

func (s *notificationStore) persistLocked() error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(s.path, data, 0o644)
}

func (s *notificationStore) listForUser(userID string, limit int32) []notificationRecord {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]notificationRecord, 0)
	for _, notification := range s.Notifications {
		if notification.UserID == userID {
			result = append(result, notification)
		}
	}

	sort.SliceStable(result, func(i, j int) bool {
		return result[i].CreatedAt > result[j].CreatedAt
	})

	if limit > 0 && int(limit) < len(result) {
		return append([]notificationRecord(nil), result[:limit]...)
	}
	return append([]notificationRecord(nil), result...)
}

func (s *notificationStore) markRead(notificationID string) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i := range s.Notifications {
		if s.Notifications[i].ID == notificationID {
			s.Notifications[i].Read = true
			return true, s.persistLocked()
		}
	}
	return false, nil
}

func (s *notificationStore) add(record notificationRecord) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.Notifications = append(s.Notifications, record)
	return s.persistLocked()
}

func (s *notificationServiceServer) GetNotifications(ctx context.Context, req *pb.GetNotificationsRequest) (*pb.GetNotificationsResponse, error) {
	if req == nil || strings.TrimSpace(req.UserId) == "" {
		return nil, status.Error(codes.InvalidArgument, "user_id is required")
	}

	records := s.store.listForUser(req.UserId, req.Limit)
	notifications := make([]*pb.Notification, 0, len(records))
	for _, record := range records {
		notifications = append(notifications, &pb.Notification{
			Id:        record.ID,
			UserId:    record.UserID,
			Type:      record.Type,
			ActorId:   record.ActorID,
			EntityId:  record.EntityID,
			CreatedAt: record.CreatedAt,
			Read:      record.Read,
		})
	}

	return &pb.GetNotificationsResponse{Notifications: notifications}, nil
}

func (s *notificationServiceServer) MarkAsRead(ctx context.Context, req *pb.MarkAsReadRequest) (*pb.MarkAsReadResponse, error) {
	if req == nil || strings.TrimSpace(req.NotificationId) == "" {
		return nil, status.Error(codes.InvalidArgument, "notification_id is required")
	}

	success, err := s.store.markRead(req.NotificationId)
	if err != nil {
		log.Printf("notification-service: failed to mark read notification_id=%s: %v", req.NotificationId, err)
		return nil, status.Error(codes.Internal, "failed to update notification")
	}
	return &pb.MarkAsReadResponse{Success: success}, nil
}

func main() {
	store, err := newNotificationStore(notificationDataPath)
	if err != nil {
		log.Fatalf("failed to initialize notification store: %v", err)
	}

	addr := "consul:8500"
	agent := consul.NewAgent(&api.Config{
		Address: addr,
	})

	serviceCfg := consul.Config{
		ServiceID:   "notification-service-1",
		ServiceName: "notification-service",
		Address:     "notification-service",
		Tags:        []string{"grpc", "notification"},
		Port:        50056,
		Check: &api.AgentServiceCheck{
			HTTP:     "http://notification-service:8080/health",
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

	lis, err := net.Listen("tcp", ":50056")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterNotificationServiceServer(s, &notificationServiceServer{store: store})
	reflection.Register(s)

	log.Println("Notification Service (gRPC) running on :50056")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
