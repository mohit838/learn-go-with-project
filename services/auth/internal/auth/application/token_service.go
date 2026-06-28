package application

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/mohit838/learn-go-with-project/internal/auth/domain"
)

type TokenService struct {
	secret     []byte
	issuer     string
	accessTTL  time.Duration
	refreshTTL time.Duration
}

func NewTokenService(secret, issuer string, accessTTL, refreshTTL time.Duration) *TokenService {
	return &TokenService{
		secret:     []byte(secret),
		issuer:     issuer,
		accessTTL:  accessTTL,
		refreshTTL: refreshTTL,
	}
}

func (s *TokenService) GeneratePair(user domain.AuthUser) (TokenResponse, error) {
	accessToken, err := s.generate(user, "access", s.accessTTL)
	if err != nil {
		return TokenResponse{}, err
	}
	refreshToken, err := s.generate(user, "refresh", s.refreshTTL)
	if err != nil {
		return TokenResponse{}, err
	}

	return TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int64(s.accessTTL.Seconds()),
	}, nil
}

func (s *TokenService) generate(user domain.AuthUser, tokenType string, ttl time.Duration) (string, error) {
	now := time.Now().UTC()
	header := map[string]string{
		"alg": "HS256",
		"typ": "JWT",
	}
	claims := map[string]any{
		"iss":        s.issuer,
		"sub":        user.PublicID,
		"tenant_id":  user.TenantPublicID,
		"tenant":     user.TenantSlug,
		"role":       user.RoleName,
		"token_type": tokenType,
		"iat":        now.Unix(),
		"exp":        now.Add(ttl).Unix(),
	}

	encodedHeader, err := encodeJWTPart(header)
	if err != nil {
		return "", err
	}
	encodedClaims, err := encodeJWTPart(claims)
	if err != nil {
		return "", err
	}

	unsigned := encodedHeader + "." + encodedClaims
	signature := sign(unsigned, s.secret)
	return unsigned + "." + signature, nil
}

func encodeJWTPart(value any) (string, error) {
	body, err := json.Marshal(value)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(body), nil
}

func sign(value string, secret []byte) string {
	mac := hmac.New(sha256.New, secret)
	mac.Write([]byte(value))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (s *TokenService) Verify(token string, expectedType string) (map[string]any, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token")
	}

	expectedSignature := sign(parts[0]+"."+parts[1], s.secret)
	if !hmac.Equal([]byte(parts[2]), []byte(expectedSignature)) {
		return nil, fmt.Errorf("invalid token signature")
	}

	claimBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}

	var claims map[string]any
	if err := json.Unmarshal(claimBytes, &claims); err != nil {
		return nil, err
	}

	if claims["token_type"] != expectedType {
		return nil, fmt.Errorf("invalid token type")
	}
	exp, ok := claims["exp"].(float64)
	if !ok || int64(exp) < time.Now().UTC().Unix() {
		return nil, fmt.Errorf("token expired")
	}

	return claims, nil
}
