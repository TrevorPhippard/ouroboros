package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"time"

	pb "ouroboros/proto/generated/notification"

	"notification/internal/consul"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/hashicorp/consul/api"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type notificationServiceServer struct {
	pb.UnimplementedNotificationServiceServer
}

// GetNotifications returns a list of notifications for a specific user
func (s *notificationServiceServer) GetNotifications(ctx context.Context, req *pb.GetNotificationsRequest) (*pb.GetNotificationsResponse, error) {
	log.Printf("Notification Service: Fetching up to %d notifications for User: %s", req.Limit, req.UserId)

	var notifications []*pb.Notification
	for i := 1; i <= int(req.Limit); i++ {
		notifications = append(notifications, &pb.Notification{
			Id:        fmt.Sprintf("notif_uuid_%d", i),
			UserId:    req.UserId,
			Type:      "FOLLOW",
			ActorId:   fmt.Sprintf("actor_%d", i),
			EntityId:  fmt.Sprintf("entity_%d", i),
			CreatedAt: time.Now().Format(time.RFC3339),
			Read:      false,
		})
	}

	return &pb.GetNotificationsResponse{
		Notifications: notifications,
	}, nil
}

// MarkAsRead updates the status of a specific notification
func (s *notificationServiceServer) MarkAsRead(ctx context.Context, req *pb.MarkAsReadRequest) (*pb.MarkAsReadResponse, error) {
	log.Printf("Notification Service: Marking notification %s as read", req.NotificationId)

	return &pb.MarkAsReadResponse{
		Success: true,
	}, nil
}

func main() {
	// Consul (Docker-safe)
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

		// HTTP health check (standardized across all services)
		Check: &api.AgentServiceCheck{
			HTTP:     "http://notification-service:8080/health",
			Interval: "10s",
			Timeout:  "2s",
		},
	}

	// Register service
	if err := agent.RegisterService(serviceCfg); err != nil {
		log.Fatalf("failed to register service: %v", err)
	}

	// HTTP server (metrics + health)
	go func() {
		http.Handle("/metrics", promhttp.Handler())

		http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("OK"))
		})

		log.Println("HTTP server running on :8080 (metrics + health)")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("failed to start HTTP server: %v", err)
		}
	}()

	// gRPC server
	lis, err := net.Listen("tcp", ":50056")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterNotificationServiceServer(s, &notificationServiceServer{})

	reflection.Register(s)

	log.Println("Notification Service (gRPC) running on :50056")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}