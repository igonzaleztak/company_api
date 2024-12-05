package token

import (
	"fmt"
	"net/http"
	"strings"
	"time"
	apierrors "xm_test/internal/api_errors"
	"xm_test/internal/conf"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the claims of a JWT token
type Claims struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	jwt.RegisteredClaims
}

// GenerateToken generates a JWT token with the given id, email
func GenerateToken(id string, email string) (string, *Claims, error) {
	claims := Claims{
		Email: email,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Minute * 10)),
			NotBefore: &jwt.NumericDate{},
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        id,
		},
	}

	accessToken := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedTokenStr, err := accessToken.SignedString([]byte(conf.GlobalConfig.JwtSecret))
	if err != nil {
		return "", nil, fmt.Errorf("failed to sign token: %w", err)
	}
	return signedTokenStr, &claims, nil
}

// ParseToken parses the given token and returns the claims
func ValidateAndParseToken(token string) (*Claims, error) {
	claims := &Claims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(conf.GlobalConfig.JwtSecret), nil
	})
	if err != nil {
		return nil, err
	}
	return claims, nil
}

// DecodeTokenFromRequest decodes the token from the Authorization header in the given request, and validates it
func DecodeTokenFromRequest(r *http.Request) (*Claims, error) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		e := apierrors.ErrTokenNotFound
		e.Message = "missing Authorization header"
		return nil, e
	}

	tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
	if tokenStr == "" {
		e := apierrors.ErrTokenNotFound
		e.Message = "missing token in Authorization header"
		return nil, e
	}

	claims, err := ValidateAndParseToken(tokenStr)
	if err != nil {
		e := apierrors.ErrInvalidToken
		e.Message = fmt.Sprintf("invalid token: %v", err)
		return nil, e
	}
	return claims, nil
}
