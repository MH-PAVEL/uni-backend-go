package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/MH-PAVEL/uni-backend-go/internal/config"
	"github.com/MH-PAVEL/uni-backend-go/internal/database"
	"github.com/MH-PAVEL/uni-backend-go/internal/routes"
)

func main() {
	// Load env
	config.LoadEnv()

	// Connect DB
	ctx, cancel := database.ConnectMongo()
	defer cancel()
	defer func() {
		if database.Client != nil {
			if err := database.Client.Disconnect(ctx); err != nil {
				log.Printf("Error disconnecting MongoDB: %v", err)
			}
		}
	}()

	// Global router
	handler := routes.RegisterRoutes()

	addr := ":8080"
	fmt.Printf("🚀 Server running on %s\n", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("Server failed: %v", err)
	}
}
