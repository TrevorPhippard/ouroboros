package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"time"

	pb "ouroboros/proto/generated/post"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"post/internal/consul"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/hashicorp/consul/api"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/segmentio/kafka-go"
)

// DB models — renamed so they do not collide with pb.Post / pb.Comment

func (DBPost) TableName() string {
	return "posts"
}

func (DBComment) TableName() string {
	return "comments"
}

func connectDB(dbURL string) *gorm.DB {
	var db *gorm.DB
	var err error

	for i := 0; i < 10; i++ {
		db, err = gorm.Open(postgres.Open(dbURL), &gorm.Config{})
		if err == nil {
			return db
		}

		log.Println("Waiting for DB...")
		time.Sleep(2 * time.Second)
	}

	log.Fatal("Failed to connect to DB:", err)
	return nil
}

func main() {
	// Consul (Docker-safe)
	addr := "consul:8500"

	agent := consul.NewAgent(&api.Config{
		Address: addr,
	})

	serviceCfg := consul.Config{
		ServiceID:   "post-service-1",
		ServiceName: "post-service",
		Address:     "post-service",
		Tags:        []string{"grpc", "post"},
		Port:        50057,

		// HTTP health check (consistent with all services)
		Check: &api.AgentServiceCheck{
			HTTP:     "http://post-service:8080/health",
			Interval: "10s",
			Timeout:  "2s",
		},
	}

	// Register service
	if err := agent.RegisterService(serviceCfg); err != nil {
		log.Fatalf("failed to register service: %v", err)
	}

	// DB setup
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL environment variable is not set")
	}

	db := connectDB(dbURL)
	log.Println("Connected to post_db")

	if err := db.AutoMigrate(&DBPost{}, &DBComment{}); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	seedDB(db)

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
	lis, err := net.Listen("tcp", ":50057")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	writer := &kafka.Writer{
		Addr:     kafka.TCP("kafka:9092"),
		Topic:    "posts.created",
		Balancer: &kafka.LeastBytes{},
	}

	// inject into service
	postService := &PostServiceServer{
		DB:     db,
		Writer: writer,
	}

	s := grpc.NewServer()
	pb.RegisterPostServiceServer(s, postService)

	reflection.Register(s)

	log.Println("Post Service running on :50057")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
