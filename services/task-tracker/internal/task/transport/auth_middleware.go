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

func RequireAuth(tokens *application.TokenService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r.Header.Get("Authorization"))
			if token == "" {
				response.Error(w, http.StatusUnauthorized, "missing bearer token", nil)
				return
			}

			user, err := tokens.VerifyAccessToken(token)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "invalid bearer token", err.Error())
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
