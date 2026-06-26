package transport

import (
	"net/http"
	"strings"

	"github.com/mohit838/learn-go-with-project/internal/auth/application"
	"github.com/mohit838/learn-go-with-project/internal/response"
)

func RequireRole(tokens *application.TokenService, role string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			token := bearerToken(r.Header.Get("Authorization"))
			if token == "" {
				response.Error(w, http.StatusUnauthorized, "missing bearer token", nil)
				return
			}

			claims, err := tokens.Verify(token, "access")
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "invalid bearer token", err.Error())
				return
			}
			if claims["role"] != role {
				response.Error(w, http.StatusForbidden, "forbidden", "required role: "+role)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
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
