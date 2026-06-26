package application

import (
	"context"
	"database/sql"
	"errors"
	"net/mail"
	"strings"

	"github.com/mohit838/learn-go-with-project/internal/auth/domain"
	"github.com/mohit838/learn-go-with-project/internal/constants"
	"github.com/mohit838/learn-go-with-project/internal/utils"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrInvalidInput       = errors.New("invalid input")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrInactiveUser       = errors.New("user is inactive")
)

type AuthService struct {
	repo   domain.Repository
	tokens *TokenService
}

func NewAuthService(repo domain.Repository, tokens *TokenService) *AuthService {
	return &AuthService{
		repo:   repo,
		tokens: tokens,
	}
}

func (s *AuthService) Register(ctx context.Context, req RegisterRequest) (AuthResponse, error) {
	req.TenantName = strings.TrimSpace(req.TenantName)
	req.TenantSlug = strings.TrimSpace(req.TenantSlug)
	roleName := constants.DefaultRoleGuest
	req.Username = strings.TrimSpace(req.Username)
	req.Email = strings.TrimSpace(req.Email)
	if err := validateRegister(req); err != nil {
		return AuthResponse{}, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return AuthResponse{}, err
	}

	user, err := s.repo.CreateTenantUser(ctx, domain.CreateUserInput{
		TenantName:   req.TenantName,
		TenantSlug:   req.TenantSlug,
		RoleName:     roleName,
		Username:     req.Username,
		Email:        req.Email,
		PasswordHash: string(hash),
	})
	if err != nil {
		return AuthResponse{}, err
	}

	return s.authResponse(user)
}

func (s *AuthService) Login(ctx context.Context, req LoginRequest) (AuthResponse, error) {
	req.TenantSlug = strings.TrimSpace(req.TenantSlug)
	req.Email = strings.TrimSpace(req.Email)
	if req.TenantSlug == "" || req.Email == "" || req.Password == "" {
		return AuthResponse{}, ErrInvalidInput
	}

	user, err := s.repo.FindUserForLogin(ctx, req.TenantSlug, req.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AuthResponse{}, ErrInvalidCredentials
		}
		return AuthResponse{}, err
	}
	if !user.IsActive {
		return AuthResponse{}, ErrInactiveUser
	}
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return AuthResponse{}, ErrInvalidCredentials
	}

	return s.authResponse(user)
}

func (s *AuthService) ListUsers(ctx context.Context, query UserListQuery) (utils.PaginatedResponse[UserResponse], error) {
	users, total, err := s.repo.ListUsers(ctx, domain.UserListFilter{
		Search:     strings.TrimSpace(query.Search),
		Role:       strings.TrimSpace(query.Role),
		TenantID:   strings.TrimSpace(query.TenantID),
		TenantSlug: strings.TrimSpace(query.TenantSlug),
		TenantName: strings.TrimSpace(query.TenantName),
		Limit:      query.PerPage,
		Offset:     query.Offset,
	})
	if err != nil {
		return utils.PaginatedResponse[UserResponse]{}, err
	}

	items := make([]UserResponse, 0, len(users))
	for _, user := range users {
		items = append(items, userResponse(user))
	}

	return utils.NewPaginatedResponse(items, utils.Pagination{
		Page:    query.Page,
		PerPage: query.PerPage,
		Offset:  query.Offset,
	}, total), nil
}

func (s *AuthService) authResponse(user domain.AuthUser) (AuthResponse, error) {
	tokens, err := s.tokens.GeneratePair(user)
	if err != nil {
		return AuthResponse{}, err
	}

	return AuthResponse{
		User:   userResponse(user),
		Tokens: tokens,
	}, nil
}

func userResponse(user domain.AuthUser) UserResponse {
	return UserResponse{
		ID:         user.PublicID,
		TenantID:   user.TenantPublicID,
		TenantName: user.TenantName,
		TenantSlug: user.TenantSlug,
		Role:       user.RoleName,
		Username:   user.Username,
		Email:      user.Email,
		IsActive:   user.IsActive,
		CreatedAt:  utils.NewAPITime(user.CreatedAt),
		UpdatedAt:  utils.NewAPITime(user.UpdatedAt),
	}
}

func validateRegister(req RegisterRequest) error {
	if req.TenantName == "" || req.TenantSlug == "" || req.Username == "" || req.Email == "" || req.Password == "" {
		return ErrInvalidInput
	}
	if _, err := mail.ParseAddress(req.Email); err != nil {
		return ErrInvalidInput
	}
	if len(req.Password) < 8 {
		return ErrInvalidInput
	}
	return nil
}
