package domain

import (
	"context"
	"time"
)

type AuthUser struct {
	ID             int64
	PublicID       string
	TenantID       int64
	TenantPublicID string
	TenantName     string
	TenantSlug     string
	RoleID         int64
	RoleName       string
	Username       string
	Email          string
	PasswordHash   string
	IsActive       bool
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type CreateUserInput struct {
	TenantName   string
	TenantSlug   string
	RoleName     string
	Username     string
	Email        string
	PasswordHash string
}

type Repository interface {
	CreateTenantUser(ctx context.Context, input CreateUserInput) (AuthUser, error)
	FindUserForLogin(ctx context.Context, tenantSlug, email string) (AuthUser, error)
	FindUserByPublicID(ctx context.Context, publicID string) (AuthUser, error)
	ListUsers(ctx context.Context, filter UserListFilter) ([]AuthUser, int64, error)
	UserStats(ctx context.Context) (UserStats, error)
}

type UserListFilter struct {
	Search     string
	Role       string
	TenantID   string
	TenantSlug string
	TenantName string
	Limit      int
	Offset     int
}

type UserStats struct {
	Total    int64
	Active   int64
	Inactive int64
	ByRole   []RoleUserCount
	ByTenant []TenantUserCount
}

type RoleUserCount struct {
	Role  string
	Count int64
}

type TenantUserCount struct {
	TenantID   string
	TenantName string
	TenantSlug string
	Count      int64
}
