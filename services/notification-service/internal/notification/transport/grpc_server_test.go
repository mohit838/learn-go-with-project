package transport

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/mohit838/learn-go-with-project/internal/notification/application"
	"github.com/mohit838/learn-go-with-project/internal/notificationrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

func TestNotificationGRPCSubmitAndStats(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := application.NewService(1, 2)
	app.Start(ctx)
	defer app.Stop()

	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	RegisterNotificationGRPCServer(server, NewNotificationGRPCServer(app))

	go func() {
		_ = server.Serve(listener)
	}()
	defer server.Stop()

	conn, err := grpc.DialContext(ctx, "bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpc.WithInsecure(),
	)
	if err != nil {
		t.Fatalf("dial grpc server: %v", err)
	}
	defer conn.Close()

	var submitted notificationrpc.NotificationResponse
	err = conn.Invoke(ctx,
		"/"+notificationrpc.ServiceName+"/SubmitNotification",
		&notificationrpc.SubmitNotificationRequest{
			UserId:    "user_1",
			TenantId:  "tenant_1",
			Role:      "admin",
			Recipient: "dev@example.com",
			Message:   "hello from grpc",
		},
		&submitted,
	)
	if err != nil {
		t.Fatalf("submit notification: %v", err)
	}
	if submitted.Id == "" || submitted.Status != "queued" {
		t.Fatalf("unexpected submit response: %+v", submitted)
	}

	time.Sleep(750 * time.Millisecond)

	var stats notificationrpc.StatsResponse
	err = conn.Invoke(ctx,
		"/"+notificationrpc.ServiceName+"/Stats",
		&notificationrpc.StatsRequest{},
		&stats,
	)
	if err != nil {
		t.Fatalf("get stats: %v", err)
	}
	if stats.Delivered != 1 {
		t.Fatalf("expected one delivered notification, got %+v", stats)
	}
}
