package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	pb "ouroboros/proto/generated/notification"

	"google.golang.org/grpc"
)

type notificationServiceServer struct {
	pb.UnimplementedNotificationServiceServer
}

// GetNotifications returns a list of notifications for a specific user
func (s *notificationServiceServer) GetNotifications(ctx context.Context, req *pb.GetNotificationsRequest) (*pb.GetNotificationsResponse, error) {
	log.Printf("Notification Service: Fetching up to %d notifications for User: %s", req.Limit, req.UserId)

	// Mocking notifications
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
	// Port 50056 selected to avoid conflicts
	lis, err := net.Listen("tcp", ":50056")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterNotificationServiceServer(s, &notificationServiceServer{})

	log.Println("Notification Service (gRPC) running on :50056")
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}