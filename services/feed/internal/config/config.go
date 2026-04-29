package config

type Config struct {
	GRPCPort   string
	HTTPPort   string
	RedisAddr  string
	KafkaAddr  string
	KafkaTopic string
	KafkaGroup string
	ConsulAddr string
}

func Load() Config {
	return Config{
		GRPCPort:   ":50055",
		HTTPPort:   ":8080",
		RedisAddr:  "redis:6379",
		KafkaAddr:  "kafka:9092",
		KafkaTopic: "posts.created",
		KafkaGroup: "feed-service",
		ConsulAddr: "consul:8500",
	}
}
