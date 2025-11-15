package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	DatabaseURL     string
	JWTSecret       string
	JWTExpiration   time.Duration
	UserManagerPort string
	GatewayPort     string
	DHTNodePort     string
	ReplicatorPort  string
}

func LoadConfig() *Config {
	return &Config{
		DatabaseURL:     getEnv("DATABASE_URL", "postgres://yourdht:yourdhtpass@localhost:5432/dht_db?sslmode=disable"),
		JWTSecret:       getEnv("JWT_SECRET", "your-secret-key-change-in-production"),
		JWTExpiration:   getDurationEnv("JWT_EXPIRATION", 1*time.Hour),
		UserManagerPort: getEnv("USERMANAGER_PORT", "8081"),
		GatewayPort:     getEnv("GATEWAY_PORT", "8080"),
		DHTNodePort:     getEnv("DHTNODE_PORT", "8082"),
		ReplicatorPort:  getEnv("REPLICATOR_PORT", "8085"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getDurationEnv(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
	}
	return defaultValue
}

func getIntEnv(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
