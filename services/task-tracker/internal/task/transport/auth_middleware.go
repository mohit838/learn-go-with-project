package transport

import (
	"context"
	"net/http"
	"strings"

	"github.com/mohit838/learn-go-with-project/internal/response"
	"github.com/mohit838/learn-go-with-project/internal/task/domain"
)

type userContextKey struct{}

// RequireGatewayAuth trusts identity headers set by APISIX/Kong.
//
// The service does not parse client JWTs here. The gateway validates the token
// and forwards only the user fields that services need for business logic.
func RequireGatewayAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if user, ok := userFromGatewayHeaders(r); ok {
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userContextKey{}, user)))
				return
			}
			response.Error(w, http.StatusUnauthorized, "missing trusted gateway identity", nil)
		})
	}
}

// UserFromContext returns the gateway identity for handlers/services.
func UserFromContext(ctx context.Context) (domain.UserContext, bool) {
	user, ok := ctx.Value(userContextKey{}).(domain.UserContext)
	return user, ok
}

// userFromGatewayHeaders reads the identity contract from the gateway.
// If any required field is missing, the request is treated as unauthenticated.
func userFromGatewayHeaders(r *http.Request) (domain.UserContext, bool) {
	user := domain.UserContext{
		UserID:     strings.TrimSpace(r.Header.Get("X-User-ID")),
		TenantID:   strings.TrimSpace(r.Header.Get("X-Tenant-ID")),
		TenantSlug: strings.TrimSpace(r.Header.Get("X-Tenant-Slug")),
		Role:       strings.TrimSpace(r.Header.Get("X-User-Role")),
	}
	return user, user.UserID != "" && user.TenantID != "" && user.Role != ""
}
