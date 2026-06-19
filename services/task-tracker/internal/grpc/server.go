package grpcserver

import (
	"context"
	"database/sql"

	"github.com/mohit838/learn-go-with-project/internal/grpc/taskv1"
	"github.com/mohit838/learn-go-with-project/internal/task"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Server struct {
	taskv1.UnimplementedTaskServiceServer
	service *task.Service
}

func NewServer(db *sql.DB) *Server {
	return &Server{service: &task.Service{Repo: &task.Repository{DB: db}}}
}

func (s *Server) GetTask(ctx context.Context, request *taskv1.GetTaskRequest) (*taskv1.GetTaskResponse, error) {
	if request.GetId() <= 0 {
		return nil, status.Error(codes.InvalidArgument, "task id must be positive")
	}
	value, err := s.service.Get(ctx, request.GetId())
	if err == task.ErrNotFound {
		return nil, status.Error(codes.NotFound, err.Error())
	}
	if err != nil {
		return nil, status.Error(codes.Internal, "get task")
	}
	return &taskv1.GetTaskResponse{
		Id:          value.ID,
		Title:       value.Title,
		Description: value.Description,
		Status:      value.Status,
		IsActive:    value.IsActive,
		CreatedAt:   value.CreatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   value.UpdatedAt.UTC().Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}
