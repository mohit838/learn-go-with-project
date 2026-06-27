package transport

import (
	"context"
	"errors"
	"time"

	"github.com/mohit838/learn-go-with-project/internal/notification/application"
	"github.com/mohit838/learn-go-with-project/internal/notification/domain"
	"github.com/mohit838/learn-go-with-project/internal/notificationrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type NotificationGRPCService interface {
	SubmitNotification(context.Context, *notificationrpc.SubmitNotificationRequest) (*notificationrpc.NotificationResponse, error)
	GetNotification(context.Context, *notificationrpc.GetNotificationRequest) (*notificationrpc.NotificationResponse, error)
	Stats(context.Context, *notificationrpc.StatsRequest) (*notificationrpc.StatsResponse, error)
}

type NotificationGRPCServer struct {
	service *application.Service
}

func NewNotificationGRPCServer(service *application.Service) *NotificationGRPCServer {
	return &NotificationGRPCServer{service: service}
}

// RegisterNotificationGRPCServer is the manual equivalent of generated
// protobuf registration code. The proto contract lives in
// proto/notification/v1/notification.proto; this descriptor wires that contract
// into grpc-go until generated pb.go files are introduced.
func RegisterNotificationGRPCServer(server *grpc.Server, handler *NotificationGRPCServer) {
	server.RegisterService(&grpc.ServiceDesc{
		ServiceName: notificationrpc.ServiceName,
		HandlerType: (*NotificationGRPCService)(nil),
		Methods: []grpc.MethodDesc{
			{MethodName: "SubmitNotification", Handler: submitNotificationHandler},
			{MethodName: "GetNotification", Handler: getNotificationHandler},
			{MethodName: "Stats", Handler: notificationStatsHandler},
		},
	}, handler)
}

// SubmitNotification lets another backend service enqueue work through gRPC.
// It uses the same application service as HTTP, so both transports share the
// same channel, worker pool, and validation rules.
func (s *NotificationGRPCServer) SubmitNotification(ctx context.Context, req *notificationrpc.SubmitNotificationRequest) (*notificationrpc.NotificationResponse, error) {
	notification, err := s.service.Submit(ctx, domain.UserContext{
		UserID:     req.UserId,
		TenantID:   req.TenantId,
		TenantSlug: req.TenantSlug,
		Role:       req.Role,
	}, application.SubmitRequest{
		Recipient: req.Recipient,
		Message:   req.Message,
	})
	if err != nil {
		return nil, grpcError(err)
	}
	return notificationResponse(notification), nil
}

func (s *NotificationGRPCServer) GetNotification(ctx context.Context, req *notificationrpc.GetNotificationRequest) (*notificationrpc.NotificationResponse, error) {
	notification, err := s.service.Find(req.Id)
	if err != nil {
		return nil, grpcError(err)
	}
	return notificationResponse(notification), nil
}

func (s *NotificationGRPCServer) Stats(context.Context, *notificationrpc.StatsRequest) (*notificationrpc.StatsResponse, error) {
	stats := s.service.Stats()
	return &notificationrpc.StatsResponse{
		Workers:     int32(stats.Workers),
		QueueSize:   int32(stats.QueueSize),
		Queued:      int32(stats.Queued),
		Sending:     int32(stats.Sending),
		Delivered:   int32(stats.Delivered),
		Failed:      int32(stats.Failed),
		JobsWaiting: int32(stats.JobsWaiting),
	}, nil
}

func submitNotificationHandler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(notificationrpc.SubmitNotificationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationGRPCService).SubmitNotification(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + notificationrpc.ServiceName + "/SubmitNotification"}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(NotificationGRPCService).SubmitNotification(ctx, req.(*notificationrpc.SubmitNotificationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func getNotificationHandler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(notificationrpc.GetNotificationRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationGRPCService).GetNotification(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + notificationrpc.ServiceName + "/GetNotification"}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(NotificationGRPCService).GetNotification(ctx, req.(*notificationrpc.GetNotificationRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func notificationStatsHandler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(notificationrpc.StatsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(NotificationGRPCService).Stats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{Server: srv, FullMethod: "/" + notificationrpc.ServiceName + "/Stats"}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(NotificationGRPCService).Stats(ctx, req.(*notificationrpc.StatsRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func notificationResponse(notification domain.Notification) *notificationrpc.NotificationResponse {
	return &notificationrpc.NotificationResponse{
		Id:          notification.ID,
		UserId:      notification.UserID,
		TenantId:    notification.TenantID,
		Recipient:   notification.Recipient,
		Message:     notification.Message,
		Status:      string(notification.Status),
		WorkerId:    int32(notification.WorkerID),
		Error:       notification.Error,
		SubmittedAt: notification.SubmittedAt.Format(time.RFC3339),
		UpdatedAt:   notification.UpdatedAt.Format(time.RFC3339),
	}
}

func grpcError(err error) error {
	switch {
	case errors.Is(err, application.ErrInvalidInput):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, application.ErrMissingUser):
		return status.Error(codes.Unauthenticated, err.Error())
	case errors.Is(err, application.ErrQueueFull):
		return status.Error(codes.ResourceExhausted, err.Error())
	case errors.Is(err, application.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, application.ErrStopped):
		return status.Error(codes.Unavailable, err.Error())
	default:
		return status.Error(codes.Internal, err.Error())
	}
}
