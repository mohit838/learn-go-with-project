package taskclient

import (
	"context"
	"net"
	"testing"

	"github.com/mohit838/learn-go-with-project/internal/grpc/taskv1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/test/bufconn"
)

type taskServer struct {
	taskv1.UnimplementedTaskServiceServer
}

func (taskServer) GetTask(_ context.Context, request *taskv1.GetTaskRequest) (*taskv1.GetTaskResponse, error) {
	return &taskv1.GetTaskResponse{Id: request.GetId(), Title: "Pay rent", Status: "pending", IsActive: true}, nil
}

func TestClientGet(t *testing.T) {
	listener := bufconn.Listen(1024 * 1024)
	server := grpc.NewServer()
	taskv1.RegisterTaskServiceServer(server, taskServer{})
	go func() { _ = server.Serve(listener) }()
	defer server.Stop()

	connection, err := grpc.DialContext(context.Background(), "bufnet", grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
		return listener.Dial()
	}), grpc.WithInsecure())
	if err != nil {
		t.Fatal(err)
	}
	defer connection.Close()

	client := &Client{connection: connection, client: taskv1.NewTaskServiceClient(connection)}
	value, err := client.Get(context.Background(), 7)
	if err != nil {
		t.Fatal(err)
	}
	if value.ID != 7 || value.Title != "Pay rent" || value.Status != "pending" || !value.IsActive {
		t.Fatalf("response = %#v", value)
	}
}
