package tokenutil

import (
	"fmt"
	"strings"
	"time"

	"github.com/amitshekhariitbhu/go-backend-clean-architecture/domain"
	jwt "github.com/golang-jwt/jwt/v4"
)

func CreateAccessToken(user *domain.User, secret string, expiry int) (string, error) {
	if user == nil {
		return "", fmt.Errorf("user is required")
	}
	if strings.TrimSpace(secret) == "" {
		return "", fmt.Errorf("token secret is not configured")
	}
	if expiry <= 0 {
		return "", fmt.Errorf("token expiry must be positive")
	}

	claims := &domain.JwtCustomClaims{
		Name:  user.Name,
		ID:    user.ID.Hex(),
		Admin: user.Admin,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * time.Duration(expiry)).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func CreateRefreshToken(user *domain.User, secret string, expiry int) (string, error) {
	if user == nil {
		return "", fmt.Errorf("user is required")
	}
	if strings.TrimSpace(secret) == "" {
		return "", fmt.Errorf("token secret is not configured")
	}
	if expiry <= 0 {
		return "", fmt.Errorf("token expiry must be positive")
	}

	claims := &domain.JwtCustomRefreshClaims{
		ID: user.ID.Hex(),
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * time.Duration(expiry)).Unix(),
			IssuedAt:  time.Now().Unix(),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secret))
}

func IsAuthorized(requestToken string, secret string) (bool, error) {
	_, err := parseAccessClaims(requestToken, secret)
	return err == nil, err
}

func ExtractIDFromToken(requestToken string, secret string) (string, error) {
	claims, err := parseAccessClaims(requestToken, secret)
	if err != nil {
		return "", err
	}
	return claims.ID, nil
}

func ExtractRoleFromToken(requestToken string, secret string) (string, error) {
	claims, err := parseAccessClaims(requestToken, secret)
	if err != nil {
		return "", err
	}
	return string(claims.Admin), nil
}

func ExtractIDFromRefreshToken(requestToken string, secret string) (string, error) {
	if strings.TrimSpace(requestToken) == "" {
		return "", fmt.Errorf("token is required")
	}
	if strings.TrimSpace(secret) == "" {
		return "", fmt.Errorf("token secret is not configured")
	}

	claims := &domain.JwtCustomRefreshClaims{}
	_, err := parseWithClaims(requestToken, secret, claims)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(claims.ID) == "" {
		return "", fmt.Errorf("refresh token subject is missing")
	}
	return claims.ID, nil
}

func parseAccessClaims(requestToken string, secret string) (*domain.JwtCustomClaims, error) {
	if strings.TrimSpace(requestToken) == "" {
		return nil, fmt.Errorf("token is required")
	}
	if strings.TrimSpace(secret) == "" {
		return nil, fmt.Errorf("token secret is not configured")
	}

	claims := &domain.JwtCustomClaims{}
	if _, err := parseWithClaims(requestToken, secret, claims); err != nil {
		return nil, err
	}
	if strings.TrimSpace(claims.ID) == "" {
		return nil, fmt.Errorf("token subject is missing")
	}
	if strings.TrimSpace(string(claims.Admin)) == "" {
		return nil, fmt.Errorf("token role is missing")
	}
	return claims, nil
}

func parseWithClaims(requestToken, secret string, claims jwt.Claims) (*jwt.Token, error) {
	parser := jwt.Parser{ValidMethods: []string{jwt.SigningMethodHS256.Alg()}}
	token, err := parser.ParseWithClaims(requestToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(secret), nil
	})
	if err != nil {
		return nil, err
	}
	if token == nil || !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}
	return token, nil
}
