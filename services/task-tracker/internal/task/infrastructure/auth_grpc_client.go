package infrastructure

import (
	"context"
	"fmt"

	"github.com/mohit838/learn-go-with-project/internal/grpcx"
	"github.com/mohit838/learn-go-with-project/internal/task/application"
	"github.com/mohit838/learn-go-with-project/internal/task/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const authServiceName = "auth.v1.AuthService"

type AuthGRPCClient struct {
	conn *grpc.ClientConn
}

func NewAuthGRPCClient(address string) (*AuthGRPCClient, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithDefaultCallOptions(grpc.CallContentSubtype(grpcx.CodecName)),
	)
	if err != nil {
		return nil, err
	}
	return &AuthGRPCClient{conn: conn}, nil
}

func (c *AuthGRPCClient) Close() error {
	return c.conn.Close()
}

func (c *AuthGRPCClient) CheckUser(ctx context.Context, user domain.UserContext) (domain.UserContext, error) {
	var res checkUserResponse
	err := c.conn.Invoke(ctx, "/"+authServiceName+"/CheckUser", checkUserRequest{
		UserID:   user.UserID,
		TenantID: user.TenantID,
	}, &res)
	if err != nil {
		return domain.UserContext{}, err
	}
	if !res.IsActive {
		return domain.UserContext{}, fmt.Errorf("user is inactive")
	}
	return domain.UserContext{
		UserID:     res.UserID,
		TenantID:   res.TenantID,
		TenantSlug: res.TenantSlug,
		Role:       res.Role,
	}, nil
}

func (c *AuthGRPCClient) UserStats(ctx context.Context) (application.UserStatsResponse, error) {
	var res userStatsResponse
	if err := c.conn.Invoke(ctx, "/"+authServiceName+"/UserStats", userStatsRequest{}, &res); err != nil {
		return application.UserStatsResponse{}, err
	}

	result := application.UserStatsResponse{
		Total:    res.Total,
		Active:   res.Active,
		Inactive: res.Inactive,
		ByRole:   make([]application.RoleUserCountResponse, 0, len(res.ByRole)),
		ByTenant: make([]application.TenantUserCountResponse, 0, len(res.ByTenant)),
	}
	for _, item := range res.ByRole {
		result.ByRole = append(result.ByRole, application.RoleUserCountResponse{
			Role:  item.Role,
			Count: item.Count,
		})
	}
	for _, item := range res.ByTenant {
		result.ByTenant = append(result.ByTenant, application.TenantUserCountResponse{
			TenantID:   item.TenantID,
			TenantName: item.TenantName,
			TenantSlug: item.TenantSlug,
			Count:      item.Count,
		})
	}

	return result, nil
}

type checkUserRequest struct {
	UserID   string `json:"user_id"`
	TenantID string `json:"tenant_id"`
}

type checkUserResponse struct {
	UserID     string `json:"user_id"`
	TenantID   string `json:"tenant_id"`
	TenantSlug string `json:"tenant_slug"`
	Role       string `json:"role"`
	IsActive   bool   `json:"is_active"`
}

type userStatsRequest struct{}

type userStatsResponse struct {
	Total    int64             `json:"total"`
	Active   int64             `json:"active"`
	Inactive int64             `json:"inactive"`
	ByRole   []roleUserCount   `json:"by_role"`
	ByTenant []tenantUserCount `json:"by_tenant"`
}

type roleUserCount struct {
	Role  string `json:"role"`
	Count int64  `json:"count"`
}

type tenantUserCount struct {
	TenantID   string `json:"tenant_id"`
	TenantName string `json:"tenant_name"`
	TenantSlug string `json:"tenant_slug"`
	Count      int64  `json:"count"`
}
