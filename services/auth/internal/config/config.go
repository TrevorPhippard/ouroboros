package config

import (
	"log"
	"os"
)

type Config struct {
	GRPCPort    string
	HTTPPort    string
	RedisAddr   string
	KafkaAddr   string
	KafkaTopic  string
	KafkaGroup  string
	ConsulAddr  string
	PostgresDSN string
	JwtSecret   []byte
}

func MustLoad() *Config {
	cfg := &Config{
		GRPCPort:   ":50055",
		HTTPPort:   ":8080",
		RedisAddr:  "redis:6379",
		KafkaAddr:  "kafka:9092",
		KafkaTopic: "auth.created",
		KafkaGroup: "auth-service",
		ConsulAddr: "consul:8500",
		// PostgresDSN: "postgresql://user:password@db:5432/auth",
		PostgresDSN: os.Getenv("DB_URL"),
		JwtSecret:   []byte(os.Getenv("JWT_SECRET")),
	}

	if cfg.PostgresDSN == "" {

		log.Fatal("missing DB_URL")
	}

	return cfg
}
