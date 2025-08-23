package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

// LoadEnv loads variables from .env (only in local/dev)
func LoadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Println("⚠️ No .env file found, relying on system ENV")
	}
}

// Get returns an env variable by key
func Get(key string) string {
	return os.Getenv(key)
}
