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
	ListUsers(ctx context.Context, filter UserListFilter) ([]AuthUser, int64, error)
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
