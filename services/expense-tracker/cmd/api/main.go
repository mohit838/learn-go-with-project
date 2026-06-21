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
	// This is the Expense Tracker API
	fmt.Println("This is the Expense Tracker API")

	// Load environment variables
	cfg, err := config.LoadConfig("./.env")
	if err != nil {
		log.Println("Error loading config:", err)
		return
	}

	fmt.Println("App Name:", cfg.AppName)
	fmt.Println("App Env:", cfg.AppEnv)
	fmt.Println("App Port:", cfg.AppPort)
	fmt.Println("Debug Mode:", cfg.AppDebug)

	// Database connection
	db, err := database.ConnectDB(cfg.DBURL)
	if err != nil {
		log.Fatalf("error connecting database: %v", err)
	}
	defer db.Close()
	log.Println("DB is connected")

	// MongoDB connection
	mongoClient, err := database.NewMongoDB(cfg.MongoURL)
	if err != nil {
		log.Fatalf("error connecting to MongoDB: %v", err)
	}
	defer func() {
		if err := mongoClient.Disconnect(context.Background()); err != nil {
			log.Fatalf("error disconnecting MongoDB: %v", err)
		}
	}()
	log.Println("MongoDB is connected")

	// Get the auth database
	mongoDB := database.GetAuthDB(mongoClient, cfg.MongoDB)

	// Redis connection
	redisClient, err := database.NewRedis(cfg.RedisURL)
	if err != nil {
		log.Fatalf("error connecting to Redis: %v", err)
	}
	defer redisClient.Close()
	log.Println("Redis is connected")

	// App routers
	handler := router.NewRouter(db, mongoDB, redisClient)

	// start the server and check port
	log.Printf("Server is running on port %s\n", cfg.AppPort)
	port := ":" + cfg.AppPort
	err = http.ListenAndServe(port, handler)
	if err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
