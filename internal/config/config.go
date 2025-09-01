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
	// Load .env file first
	LoadEnv()
	
	// Parse durations
	accessTTL, _ := time.ParseDuration(getEnv("ACCESS_TTL", "15m"))
	refreshTTL, _ := time.ParseDuration(getEnv("REFRESH_TTL", "7d"))
	
	config := &Config{
		Server: ServerConfig{
			Port: getEnv("SERVER_PORT", "8080"),
			Host: getEnv("SERVER_HOST", ""),
		},
		Database: DatabaseConfig{
			URI:  getEnv("MONGO_URI", ""),
			Name: getEnv("MONGO_DB_NAME", ""),
		},
		Auth: AuthConfig{
			JWTSecret:  getEnv("JWT_SECRET", ""),
			AccessTTL:  accessTTL,
			RefreshTTL: refreshTTL,
		},
	}

	// Validate required fields
	if config.Database.URI == "" {
		log.Fatal("MONGO_URI is required")
	}
	if config.Database.Name == "" {
		log.Fatal("MONGO_DB_NAME is required")
	}
	if config.Auth.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	AppConfig = config
	return config
}

// getEnv returns environment variable or default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
