package config

import (
	"log"
	"os"
	"strconv"
	"strings"
	"time"
)

func str(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func csv(key string, fallback []string) []string {
	var out []string
	for _, item := range strings.Split(os.Getenv(key), ",") {
		if item = strings.TrimSpace(item); item != "" {
			out = append(out, item)
		}
	}
	if len(out) == 0 {
		return fallback
	}
	return out
}

func boolean(key string, fallback bool) bool {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	parsed, err := strconv.ParseBool(v)
	if err != nil {
		log.Printf("WARNING: ignoring invalid %s=%q (want true/false)", key, v)
		return fallback
	}
	return parsed
}

func integer(key string, fallback int) int {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(v)
	if err != nil || parsed < 1 {
		log.Printf("WARNING: ignoring invalid %s=%q (want a positive integer)", key, v)
		return fallback
	}
	return parsed
}

func duration(key string, fallback time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(v)
	if err != nil || parsed <= 0 {
		log.Printf("WARNING: ignoring invalid %s=%q (want a positive duration like 3s)", key, v)
		return fallback
	}
	return parsed
}
