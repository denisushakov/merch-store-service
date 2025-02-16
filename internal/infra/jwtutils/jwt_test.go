package jwtutils

import (
	"merch-store-service/internal/infra/config"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJWTManager(t *testing.T) {
	jwtManager := NewJWTManager(&config.Config{JWTSecret: "test-secret"})

	tests := []struct {
		name        string
		userID      int
		duration    time.Duration
		expectError bool
	}{
		{"Valid token", 12345, time.Minute * 10, false},
		{"Expired token", 12345, -time.Minute, true},
		{"Zero duration (valid for 24h)", 12345, 0, false},
		{"Negative duration", 12345, -time.Hour, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			token, err := jwtManager.NewToken(tt.userID, tt.duration)
			assert.NoError(t, err, "Ошибка при создании токена")
			assert.NotEmpty(t, token, "Токен не должен быть пустым")

			claims, err := jwtManager.ValidateToken(token)
			if tt.expectError {
				assert.Error(t, err, "Ожидалась ошибка при валидации токена")
			} else {
				assert.NoError(t, err, "Ошибка при валидации токена")
				assert.Equal(t, tt.userID, claims.UserID, "Неверный userID в токене")
			}
		})
	}
}

func TestInvalidTokens(t *testing.T) {
	jwtManager := NewJWTManager(&config.Config{JWTSecret: "test-secret"})

	tests := []struct {
		name        string
		tokenString string
	}{
		{"Invalid format", "invalid.token.string"},
		{"Different signing key", func() string {
			jwtManager1 := NewJWTManager(&config.Config{JWTSecret: "secret-1"})

			token, _ := jwtManager1.NewToken(12345, time.Minute*10)
			return token
		}()},
		{"Empty token", ""},
		{"Malformed JWT", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.invalid.payload"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := jwtManager.ValidateToken(tt.tokenString)
			assert.Error(t, err, "Ожидалась ошибка при валидации токена")
		})
	}
}
