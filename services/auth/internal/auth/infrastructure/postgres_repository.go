package infrastructure

import (
	"context"
	"database/sql"
	"strings"

	"github.com/mohit838/learn-go-with-project/internal/auth/domain"
)

type AuthRepository struct {
	db *sql.DB
}

func NewAuthRepository(db *sql.DB) *AuthRepository {
	return &AuthRepository{db: db}
}

func (r *AuthRepository) CreateTenantUser(ctx context.Context, input domain.CreateUserInput) (domain.AuthUser, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.AuthUser{}, err
	}
	defer tx.Rollback()

	var user domain.AuthUser
	if err := tx.QueryRowContext(ctx, `
SELECT id, name
FROM roles
WHERE name = $1 AND is_active = TRUE`, input.RoleName).Scan(&user.RoleID, &user.RoleName); err != nil {
		return domain.AuthUser{}, err
	}

	if err := tx.QueryRowContext(ctx, `
INSERT INTO tenants (name, slug)
VALUES ($1, LOWER($2))
RETURNING id, public_id::text, slug`, input.TenantName, input.TenantSlug).Scan(
		&user.TenantID,
		&user.TenantPublicID,
		&user.TenantSlug,
	); err != nil {
		return domain.AuthUser{}, err
	}

	if err := tx.QueryRowContext(ctx, `
INSERT INTO users (tenant_id, role_id, username, email, password_hash)
VALUES ($1, $2, $3, LOWER($4), $5)
RETURNING id, public_id::text, username, email, password_hash, is_active, created_at, updated_at`,
		user.TenantID,
		user.RoleID,
		input.Username,
		input.Email,
		input.PasswordHash,
	).Scan(
		&user.ID,
		&user.PublicID,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	); err != nil {
		return domain.AuthUser{}, err
	}

	if err := tx.Commit(); err != nil {
		return domain.AuthUser{}, err
	}

	return user, nil
}

func (r *AuthRepository) FindUserForLogin(ctx context.Context, tenantSlug, email string) (domain.AuthUser, error) {
	var user domain.AuthUser
	err := r.db.QueryRowContext(ctx, `
SELECT
	u.id,
	u.public_id::text,
	u.tenant_id,
	t.public_id::text,
	t.slug,
	u.role_id,
	ro.name,
	u.username,
	u.email,
	u.password_hash,
	u.is_active,
	u.created_at,
	u.updated_at
FROM users u
JOIN tenants t ON t.id = u.tenant_id
JOIN roles ro ON ro.id = u.role_id
WHERE t.slug = LOWER($1)
	AND u.email = LOWER($2)
	AND t.is_active = TRUE
	AND ro.is_active = TRUE`, strings.TrimSpace(tenantSlug), strings.TrimSpace(email)).Scan(
		&user.ID,
		&user.PublicID,
		&user.TenantID,
		&user.TenantPublicID,
		&user.TenantSlug,
		&user.RoleID,
		&user.RoleName,
		&user.Username,
		&user.Email,
		&user.PasswordHash,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	return user, err
}
