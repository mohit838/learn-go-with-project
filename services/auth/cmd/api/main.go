package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"

	"github.com/mohit838/learn-go-with-project/internal/audit"
	"github.com/mohit838/learn-go-with-project/internal/avatar"
	"github.com/mohit838/learn-go-with-project/internal/config"
	"github.com/mohit838/learn-go-with-project/internal/database"
	appLogger "github.com/mohit838/learn-go-with-project/internal/logger"
	"github.com/mohit838/learn-go-with-project/internal/router"
	"github.com/mohit838/learn-go-with-project/internal/taskclient"
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

	cache, err := database.ConnectRedis(cfg.RedisURL)
	if err != nil {
		logger.Error("connect redis", "error", err)
		os.Exit(1)
	}
	defer cache.Close()
	logger.Info("redis connected")

	mongoClient, auditStore, err := audit.Connect(context.Background(), cfg.MongoURL, cfg.MongoDB)
	if err != nil {
		logger.Error("connect mongodb", "error", err)
		os.Exit(1)
	}
	defer mongoClient.Disconnect(context.Background())
	logger.Info("mongodb connected", "database", cfg.MongoDB)

	avatarClient, err := avatar.NewClient(cfg.AvatarAPIURL)
	if err != nil {
		logger.Error("configure avatar provider", "error", err)
		os.Exit(1)
	}
	taskClient, err := taskclient.Connect(context.Background(), cfg.TaskGRPCAddr)
	if err != nil {
		logger.Error("connect task gRPC", "error", err)
		os.Exit(1)
	}
	defer taskClient.Close()

	handler := router.NewRouter(db, logger, cache, auditStore, avatarClient, taskClient)
	port := ":" + cfg.AppPort
	logger.Info("server started", "port", cfg.AppPort, "environment", cfg.AppEnv)
	err = http.ListenAndServe(port, handler)
	if err != nil {
		logger.Error("server stopped", "error", err)
	}
}
