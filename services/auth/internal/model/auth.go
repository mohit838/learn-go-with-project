package model

import "time"

type AuthUser struct {
	ID             int64
	PublicID       string
	TenantID       int64
	TenantPublicID string
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
