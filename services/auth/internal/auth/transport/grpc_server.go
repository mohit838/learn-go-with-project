package transport

import (
	"context"
	"database/sql"

	"github.com/mohit838/learn-go-with-project/internal/auth/domain"
	"github.com/mohit838/learn-go-with-project/internal/authrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type AuthRepository interface {
	FindUserByPublicID(ctx context.Context, publicID string) (domain.AuthUser, error)
	UserStats(ctx context.Context) (domain.UserStats, error)
}

type AuthGRPCServer struct {
	repo AuthRepository
}

// AuthGRPCService mirrors proto/auth/v1/auth.proto.
//
// The request/response structs live in internal/authrpc and use protobuf wire
// tags. When protoc is added, generated pb.go files can replace that package
// without changing the application or repository layers.
type AuthGRPCService interface {
	CheckUser(context.Context, *authrpc.CheckUserRequest) (*authrpc.CheckUserResponse, error)
	UserStats(context.Context, *authrpc.UserStatsRequest) (*authrpc.UserStatsResponse, error)
}

func NewAuthGRPCServer(repo AuthRepository) *AuthGRPCServer {
	return &AuthGRPCServer{repo: repo}
}

// RegisterAuthGRPCServer manually registers the gRPC methods.
//
// Generated protobuf code normally creates this function for us. Until we add
// protoc-generated files, this small descriptor tells grpc-go which methods
// exist and which handler function should decode each request.
func RegisterAuthGRPCServer(server *grpc.Server, handler *AuthGRPCServer) {
	server.RegisterService(&grpc.ServiceDesc{
		ServiceName: authrpc.ServiceName,
		HandlerType: (*AuthGRPCService)(nil),
		Methods: []grpc.MethodDesc{
			{
				MethodName: "CheckUser",
				Handler:    checkUserHandler,
			},
			{
				MethodName: "UserStats",
				Handler:    userStatsHandler,
			},
		},
	}, handler)
}

// CheckUser is for authoritative service-to-service checks.
// Gateway auth proves the request token was valid; this method proves the user
// still exists, is active, and belongs to the tenant.
func (s *AuthGRPCServer) CheckUser(ctx context.Context, req *authrpc.CheckUserRequest) (*authrpc.CheckUserResponse, error) {
	user, err := s.repo.FindUserByPublicID(ctx, req.UserId)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, status.Error(codes.NotFound, "user not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}
	if user.TenantPublicID != req.TenantId {
		return nil, status.Error(codes.PermissionDenied, "user does not belong to tenant")
	}
	if !user.IsActive {
		return nil, status.Error(codes.PermissionDenied, "user is inactive")
	}

	return &authrpc.CheckUserResponse{
		UserId:     user.PublicID,
		TenantId:   user.TenantPublicID,
		TenantSlug: user.TenantSlug,
		Role:       user.RoleName,
		IsActive:   user.IsActive,
	}, nil
}

// UserStats powers the task dashboard. Auth owns user data, so task-tracker asks
// auth for user counts instead of querying auth tables directly.
func (s *AuthGRPCServer) UserStats(ctx context.Context, _ *authrpc.UserStatsRequest) (*authrpc.UserStatsResponse, error) {
	stats, err := s.repo.UserStats(ctx)
	if err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	res := &authrpc.UserStatsResponse{
		Total:    stats.Total,
		Active:   stats.Active,
		Inactive: stats.Inactive,
		ByRole:   make([]*authrpc.RoleUserCount, 0, len(stats.ByRole)),
		ByTenant: make([]*authrpc.TenantUserCount, 0, len(stats.ByTenant)),
	}
	for _, item := range stats.ByRole {
		res.ByRole = append(res.ByRole, &authrpc.RoleUserCount{
			Role:  item.Role,
			Count: item.Count,
		})
	}
	for _, item := range stats.ByTenant {
		res.ByTenant = append(res.ByTenant, &authrpc.TenantUserCount{
			TenantId:   item.TenantID,
			TenantName: item.TenantName,
			TenantSlug: item.TenantSlug,
			Count:      item.Count,
		})
	}

	return res, nil
}

// checkUserHandler is the manual equivalent of generated protobuf glue.
// It decodes the request body and then calls AuthGRPCServer.CheckUser.
func checkUserHandler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(authrpc.CheckUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthGRPCService).CheckUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/" + authrpc.ServiceName + "/CheckUser",
	}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(AuthGRPCService).CheckUser(ctx, req.(*authrpc.CheckUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

// userStatsHandler is the manual equivalent of generated protobuf glue for the
// UserStats method.
func userStatsHandler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(authrpc.UserStatsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthGRPCService).UserStats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/" + authrpc.ServiceName + "/UserStats",
	}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(AuthGRPCService).UserStats(ctx, req.(*authrpc.UserStatsRequest))
	}
	return interceptor(ctx, in, info, handler)
}
