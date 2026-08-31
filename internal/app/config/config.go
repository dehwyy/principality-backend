package config

import "os"

const (
	defaultMinioBaseURL = "http://localhost:9000"
	defaultServerAddr   = ":8080"
)

type Config struct {
	MinioBaseURL string
	ServerAddr   string
}

func NewConfig() *Config {
	return &Config{
		MinioBaseURL: envOrDefault("MINIO_BASE_URL", defaultMinioBaseURL),
		ServerAddr:   envOrDefault("SERVER_ADDR", defaultServerAddr),
	}
}

func envOrDefault(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
