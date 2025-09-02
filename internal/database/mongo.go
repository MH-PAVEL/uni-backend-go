package database

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/MH-PAVEL/uni-backend-go/internal/config"
	"go.mongodb.org/mongo-driver/bson"
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

// CreateIndexes creates necessary database indexes
func CreateIndexes() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create unique index on NID field
	usersCollection := GetCollection(DbName(), UsersCollection)
	
	// Create unique index on NID (National ID)
	_, err := usersCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "nid", Value: 1}},
		Options: options.Index().SetUnique(true).SetSparse(true),
	})
	if err != nil {
		return fmt.Errorf("failed to create NID index: %w", err)
	}

	// Create unique index on email
	_, err = usersCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "email", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("failed to create email index: %w", err)
	}

	// Create unique index on phone
	_, err = usersCollection.Indexes().CreateOne(ctx, mongo.IndexModel{
		Keys: bson.D{{Key: "phone", Value: 1}},
		Options: options.Index().SetUnique(true),
	})
	if err != nil {
		return fmt.Errorf("failed to create phone index: %w", err)
	}

	return nil
}
