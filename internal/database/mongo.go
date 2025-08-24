package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/MH-PAVEL/uni-backend-go/internal/config"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var Client *mongo.Client

// ConnectMongo initializes the MongoDB client and returns a context + cancel function
func ConnectMongo() (context.Context, context.CancelFunc) {
	cfg := config.AppConfig
	if cfg == nil {
		log.Fatal("Configuration not loaded")
	}

	// Root context for the entire DB session
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(cfg.Database.URI))
	if err != nil {
		log.Fatalf("Failed to connect MongoDB: %v", err)
	}

	// Check connection
	if err = client.Ping(ctx, nil); err != nil {
		log.Fatalf("Failed to ping MongoDB: %v", err)
	}

	Client = client
	fmt.Println("✅ Connected to MongoDB")

	return ctx, cancel
}

// GetCollection returns a MongoDB collection instance
func GetCollection(dbName, collName string) *mongo.Collection {
	return Client.Database(dbName).Collection(collName)
}
