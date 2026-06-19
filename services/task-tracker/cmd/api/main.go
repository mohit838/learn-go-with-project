package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/mohit838/learn-go-with-project/internal/config"
)

func main() {
	// This is the Task Tracker API
	fmt.Println("This is the Task Tracker API")

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

	// Initialize database connection (later)
	// Initialize Redis connection (later)

	// Set up routes and handlers
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Task Tracker API is running!"))
	})

	// Start the server
	appPort := cfg.AppPort
	if appPort == "" {
		log.Fatal("Port is not specified")
	}

	fmt.Printf("Starting server on port %s...\n", appPort)
	if err := http.ListenAndServe(":"+appPort, r); err != nil {
		log.Fatal("Server failed:", err)
	}
}
