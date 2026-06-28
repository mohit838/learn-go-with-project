package transport

import (
	"context"
	"net"
	"testing"

	"github.com/mohit838/learn-go-with-project/internal/auth/domain"
	"github.com/mohit838/learn-go-with-project/internal/authrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
)

func TestAuthGRPCServerUsesDefaultProtoCodec(t *testing.T) {
	t.Parallel()

	server := grpc.NewServer()
	RegisterAuthGRPCServer(server, NewAuthGRPCServer(fakeAuthRepo{}))

	listener := bufconn.Listen(1024 * 1024)
	go func() {
		if err := server.Serve(listener); err != nil {
			t.Errorf("serve: %v", err)
		}
	}()
	defer server.Stop()

	conn, err := grpc.DialContext(
		context.Background(),
		"bufnet",
		grpc.WithContextDialer(func(context.Context, string) (net.Conn, error) {
			return listener.Dial()
		}),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer conn.Close()

	var checkUser authrpc.CheckUserResponse
	if err := conn.Invoke(context.Background(), "/"+authrpc.ServiceName+"/CheckUser", &authrpc.CheckUserRequest{
		UserId:   "user_01",
		TenantId: "tenant_01",
	}, &checkUser); err != nil {
		t.Fatalf("CheckUser invoke: %v", err)
	}
	if checkUser.Role != "superadmin" || !checkUser.IsActive {
		t.Fatalf("unexpected CheckUser response: %+v", checkUser)
	}

	var stats authrpc.UserStatsResponse
	if err := conn.Invoke(context.Background(), "/"+authrpc.ServiceName+"/UserStats", &authrpc.UserStatsRequest{}, &stats); err != nil {
		t.Fatalf("UserStats invoke: %v", err)
	}
	if stats.Total != 2 || stats.ByRole[0].Role != "superadmin" {
		t.Fatalf("unexpected UserStats response: %+v", stats)
	}
}

type fakeAuthRepo struct{}

func (fakeAuthRepo) FindUserByPublicID(context.Context, string) (domain.AuthUser, error) {
	return domain.AuthUser{
		PublicID:       "user_01",
		TenantPublicID: "tenant_01",
		TenantSlug:     "default",
		RoleName:       "superadmin",
		IsActive:       true,
	}, nil
}

func (fakeAuthRepo) UserStats(context.Context) (domain.UserStats, error) {
	return domain.UserStats{
		Total:    2,
		Active:   2,
		Inactive: 0,
		ByRole: []domain.RoleUserCount{
			{Role: "superadmin", Count: 1},
			{Role: "staff", Count: 1},
		},
		ByTenant: []domain.TenantUserCount{
			{TenantID: "tenant_01", TenantName: "Default", TenantSlug: "default", Count: 2},
		},
	}, nil
}
