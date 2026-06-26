package application

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mohit838/learn-go-with-project/internal/task/domain"
)

type TokenService struct {
	secret []byte
	issuer string
}

func NewTokenService(secret, issuer string) *TokenService {
	return &TokenService{
		secret: []byte(secret),
		issuer: issuer,
	}
}

func (s *TokenService) VerifyAccessToken(token string) (domain.UserContext, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return domain.UserContext{}, fmt.Errorf("invalid token")
	}

	expectedSignature := sign(parts[0]+"."+parts[1], s.secret)
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSignature)) {
		return domain.UserContext{}, fmt.Errorf("invalid token signature")
	}

	claimBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return domain.UserContext{}, err
	}

	var claims map[string]any
	if err := json.Unmarshal(claimBytes, &claims); err != nil {
		return domain.UserContext{}, err
	}
	if claims["iss"] != s.issuer {
		return domain.UserContext{}, fmt.Errorf("invalid token issuer")
	}
	if claims["token_type"] != "access" {
		return domain.UserContext{}, fmt.Errorf("invalid token type")
	}
	exp, ok := claims["exp"].(float64)
	if !ok || int64(exp) < time.Now().UTC().Unix() {
		return domain.UserContext{}, fmt.Errorf("token expired")
	}

	return domain.UserContext{
		UserID:     claimString(claims, "sub"),
		TenantID:   claimString(claims, "tenant_id"),
		TenantSlug: claimString(claims, "tenant"),
		Role:       claimString(claims, "role"),
	}, nil
}

func sign(value string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func claimString(claims map[string]any, key string) string {
	value, _ := claims[key].(string)
	return value
}
