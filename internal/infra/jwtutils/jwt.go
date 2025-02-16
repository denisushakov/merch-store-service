package jwtutils

import (
	"errors"
	"fmt"
	"merch-store-service/internal/infra/config"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var (
	ErrInvalidToken       = errors.New("Unauthorized")
	ErrInvalidSigning     = errors.New("Unauthorized")
	ErrInvalidTokenClaims = errors.New("Unauthorized")
)

type JWTManager struct {
	SecretKey string
}

func NewJWTManager(config *config.Config) *JWTManager {
	return &JWTManager{
		SecretKey: config.JWTSecret,
	}
}

type Claims struct {
	UserID int `json:"user_id"`
	jwt.RegisteredClaims
}

func (j *JWTManager) NewToken(UserID int, duration time.Duration) (string, error) {
	expirationTime := time.Now().Add(duration)
	if duration == 0 {
		expirationTime = time.Now().Add(time.Hour * 24)
	}

	claims := Claims{
		UserID: UserID,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	return token.SignedString([]byte(j.SecretKey))
}

// ValidateToken проверяет JWT и возвращает payload
func (j *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	parsedToken, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidSigning
		}
		return []byte(j.SecretKey), nil
	})

	if err != nil {
		return nil, fmt.Errorf("token parsing failed: %w", err)
	}

	claims, ok := parsedToken.Claims.(*Claims)
	if !ok || !parsedToken.Valid {
		return nil, ErrInvalidTokenClaims
	}

	if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, ErrInvalidToken
	}

	return claims, nil
}
