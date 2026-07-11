package infrastructure

import (
	"context"
	"fmt"

	"github.com/mohit838/learn-go-with-project/internal/authrpc"
	"github.com/mohit838/learn-go-with-project/internal/task/application"
	"github.com/mohit838/learn-go-with-project/internal/task/domain"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AuthGRPCClient is task-tracker's private client for auth service.
//
// Normal HTTP requests trust gateway headers after Kong validates JWTs.
// This client is for service-to-service questions that only auth can answer,
// such as "give me user dashboard counts" or stricter future user checks.
type AuthGRPCClient struct {
	conn *grpc.ClientConn
}

func NewAuthGRPCClient(address string) (*AuthGRPCClient, error) {
	conn, err := grpc.NewClient(
		address,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}
	return &AuthGRPCClient{conn: conn}, nil
}

func (c *AuthGRPCClient) Close() error {
	return c.conn.Close()
}

// CheckUser asks auth for the current truth about a user.
//
// We do not call this on every request because the gateway already validates
// access tokens. Keep this for stricter business rules, for example before a
// sensitive cross-service action.
func (c *AuthGRPCClient) CheckUser(ctx context.Context, user domain.UserContext) (domain.UserContext, error) {
	var res authrpc.CheckUserResponse
	err := c.conn.Invoke(ctx, "/"+authrpc.ServiceName+"/CheckUser", &authrpc.CheckUserRequest{
		UserId:   user.UserID,
		TenantId: user.TenantID,
	}, &res)
	if err != nil {
		return domain.UserContext{}, err
	}
	if !res.IsActive {
		return domain.UserContext{}, fmt.Errorf("user is inactive")
	}
	return domain.UserContext{
		UserID:     res.UserId,
		TenantID:   res.TenantId,
		TenantSlug: res.TenantSlug,
		Role:       res.Role,
	}, nil
}

// UserStats asks auth for user aggregates used by the superadmin dashboard.
// Task-tracker owns task data; auth owns user data. gRPC lets each service keep
// its own database boundary.
func (c *AuthGRPCClient) UserStats(ctx context.Context) (application.UserStatsResponse, error) {
	var res authrpc.UserStatsResponse
	if err := c.conn.Invoke(ctx, "/"+authrpc.ServiceName+"/UserStats", &authrpc.UserStatsRequest{}, &res); err != nil {
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
			TenantID:   item.TenantId,
			TenantName: item.TenantName,
			TenantSlug: item.TenantSlug,
			Count:      item.Count,
		})
	}

	return result, nil
}
