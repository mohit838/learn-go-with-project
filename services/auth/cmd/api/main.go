package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/mohit838/learn-go-with-project/internal/config"
	"github.com/mohit838/learn-go-with-project/internal/database"
	"github.com/mohit838/learn-go-with-project/internal/router"
)

func main() {
	// ========================
	// Auth Service API Entry Point
	// ========================
	fmt.Println("\n=== Auth Service API ===")

	// Load environment variables from .env file
	cfg, err := config.LoadConfig("./.env")
	if err != nil {
		log.Println("Error loading config:", err)
		return
	}

	// Log startup configuration
	fmt.Printf("App Name: %s | Env: %s | Port: %s | Debug: %v\n\n",
		cfg.AppName, cfg.AppEnv, cfg.AppPort, cfg.AppDebug)

	// ========================
	// Database Connections
	// ========================

	// PostgreSQL connection for primary data storage
	db, err := database.ConnectDB(cfg.DBURL)
	if err != nil {
		log.Fatalf("error connecting database: %v", err)
	}
	defer db.Close()
	log.Println(">>-->> PostgreSQL connected")

	// MongoDB connection for audit/event logging
	// Database: auth_log_db | Collection: auth_logs
	mongoClient, err := database.NewMongoDB(cfg.MongoURL)
	if err != nil {
		log.Fatalf("error connecting to MongoDB: %v", err)
	}
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Fatalf("error disconnecting MongoDB: %v", err)
		}
	}()
	log.Println(">>-->> MongoDB connected")

	// Get MongoDB database instance
	mongoDB := database.GetAuthDB(mongoClient, cfg.MongoDB)

	// Redis connection for caching (Database 0)
	// Used for session storage and short-lived data caching
	redisClient, err := database.NewRedis(cfg.RedisURL)
	if err != nil {
		log.Fatalf("error connecting to Redis: %v", err)
	}
	defer redisClient.Close()
	log.Println(">>-->> Redis connected")

	// ========================
	// Initialize Router & Start Server
	// ========================

	// Initialize HTTP router with all middleware
	handler := router.NewRouter(db, mongoDB, redisClient)

	// Start HTTP server on configured port
	log.Printf("Server starting on port %s...\n", cfg.AppPort)
	port := ":" + cfg.AppPort
	err = http.ListenAndServe(port, handler)
	if err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
