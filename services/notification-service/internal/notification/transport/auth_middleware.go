package transport

import (
	"context"
	"net/http"
	"strings"

	"github.com/mohit838/learn-go-with-project/internal/notification/domain"
	"github.com/mohit838/learn-go-with-project/internal/response"
)

type userContextKey struct{}

func RequireGatewayAuth() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			user := domain.UserContext{
				UserID:     strings.TrimSpace(r.Header.Get("X-User-ID")),
				TenantID:   strings.TrimSpace(r.Header.Get("X-Tenant-ID")),
				TenantSlug: strings.TrimSpace(r.Header.Get("X-Tenant-Slug")),
				Role:       strings.TrimSpace(r.Header.Get("X-User-Role")),
			}
			if user.UserID == "" || user.TenantID == "" || user.Role == "" {
				response.Error(w, http.StatusUnauthorized, "missing trusted gateway identity", nil)
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userContextKey{}, user)))
		})
	}
}

func UserFromContext(ctx context.Context) (domain.UserContext, bool) {
	user, ok := ctx.Value(userContextKey{}).(domain.UserContext)
	return user, ok
}
