package main

import (
	"log"
	"net"
	"net/http"
	"os"
	"time"

	pb "ouroboros/proto/generated/profile"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"profile/internal/consul"

	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"

	"github.com/hashicorp/consul/api"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

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
	// ✅ Consul (Docker-safe DNS)
	addr := "consul:8500"

	agent := consul.NewAgent(&api.Config{
		Address: addr,
	})

	serviceCfg := consul.Config{
		ServiceID:   "profile-service-1",
		ServiceName: "profile-service",
		Address:     "profile-service",
		Tags:        []string{"grpc", "profile"},
		Port:        50058,

		// ✅ HTTP health check (standardized across system)
		Check: &api.AgentServiceCheck{
			HTTP:     "http://profile-service:8080/health",
			Interval: "10s",
			Timeout:  "2s",
		},
	}

	// ✅ Register service with Consul
	if err := agent.RegisterService(serviceCfg); err != nil {
		log.Fatalf("failed to register service: %v", err)
	}

	// DB setup
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL environment variable is not set")
	}

	db := connectDB(dbURL)

	log.Println("Connected to profile_db")

	if err := db.AutoMigrate(&Profile{}); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	seedDB(db)

	// ✅ HTTP server (metrics + health)
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
	lis, err := net.Listen("tcp", ":50058")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer()
	pb.RegisterProfileServiceServer(s, &profileServiceServer{})

	reflection.Register(s)

	log.Println("Profile Service (gRPC) running on :50058")

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}