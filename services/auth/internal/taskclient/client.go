package taskclient

import (
	"context"
	"errors"
	"time"

	"github.com/mohit838/learn-go-with-project/internal/grpc/taskv1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

var ErrDisabled = errors.New("task gRPC client is not configured")

type Response struct {
	ID          int64  `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Status      string `json:"status"`
	IsActive    bool   `json:"isActive"`
	CreatedAt   string `json:"createdAt"`
	UpdatedAt   string `json:"updatedAt"`
}

type Client struct {
	connection *grpc.ClientConn
	client     taskv1.TaskServiceClient
}

func Connect(ctx context.Context, address string) (*Client, error) {
	if address == "" {
		return nil, nil
	}
	connection, err := grpc.DialContext(ctx, address, grpc.WithTransportCredentials(insecure.NewCredentials()), grpc.WithBlock())
	if err != nil {
		return nil, err
	}
	return &Client{connection: connection, client: taskv1.NewTaskServiceClient(connection)}, nil
}

func (c *Client) Close() error {
	if c == nil {
		return nil
	}
	return c.connection.Close()
}

func (c *Client) Get(ctx context.Context, id int64) (Response, error) {
	if c == nil {
		return Response{}, ErrDisabled
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	value, err := c.client.GetTask(ctx, &taskv1.GetTaskRequest{Id: id})
	if err != nil {
		return Response{}, err
	}
	return Response{ID: value.GetId(), Title: value.GetTitle(), Description: value.GetDescription(), Status: value.GetStatus(), IsActive: value.GetIsActive(), CreatedAt: value.GetCreatedAt(), UpdatedAt: value.GetUpdatedAt()}, nil
}
