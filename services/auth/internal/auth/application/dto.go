package application

import "github.com/mohit838/learn-go-with-project/internal/utils"

type RegisterRequest struct {
	TenantName string `json:"tenant_name"`
	TenantSlug string `json:"tenant_slug"`
	RoleName   string `json:"role_name,omitempty"`
	Username   string `json:"username"`
	Email      string `json:"email"`
	Password   string `json:"password"`
}

type LoginRequest struct {
	TenantSlug string `json:"tenant_slug"`
	Email      string `json:"email"`
	Password   string `json:"password"`
}

type AuthResponse struct {
	User   UserResponse  `json:"user"`
	Tokens TokenResponse `json:"tokens"`
}

type UserResponse struct {
	ID         string        `json:"id"`
	TenantID   string        `json:"tenant_id"`
	TenantSlug string        `json:"tenant_slug"`
	Role       string        `json:"role"`
	Username   string        `json:"username"`
	Email      string        `json:"email"`
	IsActive   bool          `json:"is_active"`
	CreatedAt  utils.APITime `json:"created_at"`
	UpdatedAt  utils.APITime `json:"updated_at"`
}

type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
}
