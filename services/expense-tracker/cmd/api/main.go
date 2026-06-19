package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/mohit838/learn-go-with-project/internal/audit"
	"github.com/mohit838/learn-go-with-project/internal/config"
	"github.com/mohit838/learn-go-with-project/internal/database"
	appLogger "github.com/mohit838/learn-go-with-project/internal/logger"
	"github.com/mohit838/learn-go-with-project/internal/router"
)

func main() {
	cfg, err := config.LoadConfig("./.env")
	if err != nil {
		slog.Error("load config", "error", err)
		os.Exit(1)
	}

	logger := appLogger.New(cfg.LogLevel).With("service", cfg.AppName)

	db, err := database.ConnectDB(cfg.DBURL)
	if err != nil {
		logger.Error("connect database", "error", err)
		os.Exit(1)
	}
	defer db.Close()
	logger.Info("database connected")

	mongoClient, auditStore, err := audit.Connect(context.Background(), cfg.MongoURL, cfg.MongoDB)
	if err != nil {
		logger.Error("connect mongodb", "error", err)
		os.Exit(1)
	}
	defer mongoClient.Disconnect(context.Background())
	logger.Info("mongodb connected", "database", cfg.MongoDB)

	handler := router.NewRouter(db, logger, auditStore)
	port := ":" + cfg.AppPort
	logger.Info("server started", "port", cfg.AppPort, "environment", cfg.AppEnv)
	err = http.ListenAndServe(port, handler)
	if err != nil {
		logger.Error("server stopped", "error", err)
	}
}
