package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/mohit838/learn-go-with-project/internal/config"
	"github.com/mohit838/learn-go-with-project/internal/notification/application"
	"github.com/mohit838/learn-go-with-project/internal/notification/transport"
	"github.com/mohit838/learn-go-with-project/internal/router"
	"google.golang.org/grpc"
)

func main() {
	fmt.Println("\n=== Notification Service API ===")

	cfg, err := config.LoadConfig("./.env")
	if err != nil {
		log.Fatalf("load config: %v", err)
	}
	fmt.Printf("App Name: %s | Env: %s | Port: %s | Debug: %v\n\n",
		cfg.AppName, cfg.AppEnv, cfg.AppPort, cfg.AppDebug)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	notificationService := application.NewService(cfg.Workers, cfg.QueueSize)
	notificationService.Start(ctx)
	defer notificationService.Stop()

	grpcServer := grpc.NewServer()
	transport.RegisterNotificationGRPCServer(grpcServer, transport.NewNotificationGRPCServer(notificationService))

	server := &http.Server{
		Addr:              ":" + cfg.AppPort,
		Handler:           router.NewRouter(notificationService),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = server.Shutdown(shutdownCtx)
		grpcServer.GracefulStop()
	}()

	errs := make(chan error, 2)

	go func() {
		listener, err := net.Listen("tcp", ":"+cfg.GRPCPort)
		if err != nil {
			errs <- fmt.Errorf("grpc listen: %w", err)
			return
		}
		log.Printf("gRPC server starting on port %s...\n", cfg.GRPCPort)
		if err := grpcServer.Serve(listener); err != nil {
			errs <- fmt.Errorf("grpc server: %w", err)
		}
	}()

	go func() {
		log.Printf("HTTP server starting on port %s...\n", cfg.AppPort)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			errs <- fmt.Errorf("http server: %w", err)
		}
	}()

	select {
	case <-ctx.Done():
	case err := <-errs:
		log.Fatalf("server failed: %v", err)
	}
}
