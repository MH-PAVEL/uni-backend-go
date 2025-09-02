package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/MH-PAVEL/uni-backend-go/internal/config"
	"github.com/MH-PAVEL/uni-backend-go/internal/database"
	_ "github.com/MH-PAVEL/uni-backend-go/internal/docs"

	"github.com/MH-PAVEL/uni-backend-go/internal/routes"
)

// @title           Uni Backend API
// @version         1.0
// @description     University backend API.
// @BasePath        /
// @schemes         http
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Use: "Bearer <access_token>"

func main() {
	// Load env and config
	config.LoadEnv()
	cfg := config.LoadConfig()

	// Connect DB
	_, dbCancel := database.ConnectMongo()
	defer dbCancel()

	// Create database indexes
	if err := database.CreateIndexes(); err != nil {
		log.Printf("Warning: Failed to create database indexes: %v", err)
	}

	// Global router
	handler := routes.RegisterRoutes()



	// server address
	addr := fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	
	// Create HTTP server
	server := &http.Server{
		Addr:    addr,
		Handler: handler,
		ReadTimeout: 10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		fmt.Printf("🚀 Server running on %s\n", addr)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("🛑 Shutting down server...")

	// Create shutdown context with 30s timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	// Shutdown server
	if err := server.Shutdown(shutdownCtx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	// Disconnect database
	if database.Client != nil {
		if err := database.Client.Disconnect(shutdownCtx); err != nil {
			log.Printf("Error disconnecting MongoDB: %v", err)
		}
	}

	log.Println("✅ Server exited")
}
