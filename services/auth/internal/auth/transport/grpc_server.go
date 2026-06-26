package transport

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/mohit838/learn-go-with-project/internal/auth/domain"
	"google.golang.org/grpc"
)

const AuthServiceName = "auth.v1.AuthService"

type AuthRepository interface {
	FindUserByPublicID(ctx context.Context, publicID string) (domain.AuthUser, error)
	UserStats(ctx context.Context) (domain.UserStats, error)
}

type AuthGRPCServer struct {
	repo AuthRepository
}

type AuthGRPCService interface {
	CheckUser(context.Context, *CheckUserRequest) (*CheckUserResponse, error)
	UserStats(context.Context, *UserStatsRequest) (*UserStatsResponse, error)
}

func NewAuthGRPCServer(repo AuthRepository) *AuthGRPCServer {
	return &AuthGRPCServer{repo: repo}
}

func RegisterAuthGRPCServer(server *grpc.Server, handler *AuthGRPCServer) {
	server.RegisterService(&grpc.ServiceDesc{
		ServiceName: AuthServiceName,
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

type CheckUserRequest struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
}

type CheckUserResponse struct {
	UserID     string `json:"user_id"`
	TenantID   string `json:"tenant_id"`
	TenantSlug string `json:"tenant_slug"`
	Role       string `json:"role"`
	IsActive   bool   `json:"is_active"`
}

type UserStatsRequest struct{}

type UserStatsResponse struct {
	Total    int64             `json:"total"`
	Active   int64             `json:"active"`
	Inactive int64             `json:"inactive"`
	ByRole   []RoleUserCount   `json:"by_role"`
	ByTenant []TenantUserCount `json:"by_tenant"`
}

type RoleUserCount struct {
	Role  string `json:"role"`
	Count int64  `json:"count"`
}

type TenantUserCount struct {
	TenantID   string `json:"tenant_id"`
	TenantName string `json:"tenant_name"`
	TenantSlug string `json:"tenant_slug"`
	Count      int64  `json:"count"`
}

func (s *AuthGRPCServer) CheckUser(ctx context.Context, req *CheckUserRequest) (*CheckUserResponse, error) {
	user, err := s.repo.FindUserByPublicID(ctx, req.UserID)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("user not found")
		}
		return nil, err
	}
	if user.TenantPublicID != req.TenantID {
		return nil, fmt.Errorf("user does not belong to tenant")
	}
	if !user.IsActive {
		return nil, fmt.Errorf("user is inactive")
	}

	return &CheckUserResponse{
		UserID:     user.PublicID,
		TenantID:   user.TenantPublicID,
		TenantSlug: user.TenantSlug,
		Role:       user.RoleName,
		IsActive:   user.IsActive,
	}, nil
}

func (s *AuthGRPCServer) UserStats(ctx context.Context, _ *UserStatsRequest) (*UserStatsResponse, error) {
	stats, err := s.repo.UserStats(ctx)
	if err != nil {
		return nil, err
	}

	res := &UserStatsResponse{
		Total:    stats.Total,
		Active:   stats.Active,
		Inactive: stats.Inactive,
		ByRole:   make([]RoleUserCount, 0, len(stats.ByRole)),
		ByTenant: make([]TenantUserCount, 0, len(stats.ByTenant)),
	}
	for _, item := range stats.ByRole {
		res.ByRole = append(res.ByRole, RoleUserCount{
			Role:  item.Role,
			Count: item.Count,
		})
	}
	for _, item := range stats.ByTenant {
		res.ByTenant = append(res.ByTenant, TenantUserCount{
			TenantID:   item.TenantID,
			TenantName: item.TenantName,
			TenantSlug: item.TenantSlug,
			Count:      item.Count,
		})
	}

	return res, nil
}

func checkUserHandler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(CheckUserRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthGRPCService).CheckUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/" + AuthServiceName + "/CheckUser",
	}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(AuthGRPCService).CheckUser(ctx, req.(*CheckUserRequest))
	}
	return interceptor(ctx, in, info, handler)
}

func userStatsHandler(srv any, ctx context.Context, dec func(any) error, interceptor grpc.UnaryServerInterceptor) (any, error) {
	in := new(UserStatsRequest)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(AuthGRPCService).UserStats(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/" + AuthServiceName + "/UserStats",
	}
	handler := func(ctx context.Context, req any) (any, error) {
		return srv.(AuthGRPCService).UserStats(ctx, req.(*UserStatsRequest))
	}
	return interceptor(ctx, in, info, handler)
}
