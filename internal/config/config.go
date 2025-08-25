package config

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
)


type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Auth     AuthConfig
}


type ServerConfig struct {
	Port string
	Host string
}


type DatabaseConfig struct {
	URI  string
	Name string
}

type AuthConfig struct {
	JWTSecret   string
	AccessTTL   time.Duration
	RefreshTTL  time.Duration
}

var AppConfig *Config

// LoadEnv loads variables from .env (only in local/dev)
func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ No .env file found, relying on system ENV")
	}
}

// LoadConfig loads and validates all configuration
func LoadConfig() *Config {
	config := &Config{
		Server: ServerConfig{
			Port: getEnvWithDefault("SERVER_PORT", "8080"),
			Host: getEnvWithDefault("SERVER_HOST", ""),
		},
		Database: DatabaseConfig{
			URI:  getEnvRequired("MONGO_URI"),
			Name: getEnvRequired("MONGO_DB_NAME"),
		},
		Auth: AuthConfig{
			JWTSecret:  getEnvRequired("JWT_SECRET"),
			AccessTTL:  getDurationEnvWithDefault("ACCESS_TTL", 7*24*time.Hour),
			RefreshTTL: getDurationEnvWithDefault("REFRESH_TTL", 30*24*time.Hour),
		},
	}

	AppConfig = config
	return config
}

// Get returns an env variable by key
func Get(key string) string {
	return os.Getenv(key)
}

// getEnvWithDefault returns environment variable or default value
func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getEnvRequired returns environment variable or panics if not set
func getEnvRequired(key string) string {
	value := os.Getenv(key)
	if value == "" {
		log.Fatalf("Required environment variable %s is not set", key)
	}
	return value
}

// getDurationEnvWithDefault returns duration from environment variable or default value
func getDurationEnvWithDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if duration, err := time.ParseDuration(value); err == nil {
			return duration
		}
		log.Printf("Invalid duration format for %s, using default", key)
	}
	return defaultValue
}
