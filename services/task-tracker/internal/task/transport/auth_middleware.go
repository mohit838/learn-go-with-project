package transport

import (
	"context"
	"net/http"
	"strings"

	"github.com/mohit838/learn-go-with-project/internal/response"
	"github.com/mohit838/learn-go-with-project/internal/task/application"
	"github.com/mohit838/learn-go-with-project/internal/task/domain"
)

type userContextKey struct{}

type AuthoritativeUserChecker interface {
	CheckUser(ctx context.Context, user domain.UserContext) (domain.UserContext, error)
}

func RequireAuth(tokens *application.TokenService, checker AuthoritativeUserChecker) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if user, ok := userFromGatewayHeaders(r); ok {
				next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userContextKey{}, user)))
				return
			}

			token := bearerToken(r.Header.Get("Authorization"))
			if token == "" {
				response.Error(w, http.StatusUnauthorized, "missing gateway identity or bearer token", nil)
				return
			}

			user, err := tokens.VerifyAccessToken(token)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "invalid bearer token", err.Error())
				return
			}
			if checker != nil {
				user, err = checker.CheckUser(r.Context(), user)
				if err != nil {
					response.Error(w, http.StatusUnauthorized, "user is not authorized by auth service", err.Error())
					return
				}
			}

			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), userContextKey{}, user)))
		})
	}
}

func UserFromContext(ctx context.Context) (domain.UserContext, bool) {
	user, ok := ctx.Value(userContextKey{}).(domain.UserContext)
	return user, ok
}

func bearerToken(header string) string {
	if header == "" {
		return ""
	}
	parts := strings.SplitN(header, " ", 2)
	if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
		return ""
	}
	return strings.TrimSpace(parts[1])
}

func userFromGatewayHeaders(r *http.Request) (domain.UserContext, bool) {
	user := domain.UserContext{
		UserID:     strings.TrimSpace(r.Header.Get("X-User-ID")),
		TenantID:   strings.TrimSpace(r.Header.Get("X-Tenant-ID")),
		TenantSlug: strings.TrimSpace(r.Header.Get("X-Tenant-Slug")),
		Role:       strings.TrimSpace(r.Header.Get("X-User-Role")),
	}
	return user, user.UserID != "" && user.TenantID != "" && user.Role != ""
}
