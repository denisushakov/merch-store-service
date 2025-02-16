package middleware

import (
	"context"
	"merch-store-service/internal/infra/config"
	"merch-store-service/internal/infra/jwtutils"
	"merch-store-service/pkg/ctxkeys"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type mockHandler struct {
	called bool
	ctx    context.Context
}

func (m *mockHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	m.called = true
	m.ctx = r.Context()
}

func TestAuthMiddleware(t *testing.T) {
	jwtManager := jwtutils.NewJWTManager(&config.Config{JWTSecret: "test-secret"})
	middleware := NewAuthMiddleware(jwtManager)

	validToken, _ := jwtManager.NewToken(12345, time.Minute*10)
	expiredToken, _ := jwtManager.NewToken(12345, -time.Minute)

	tests := []struct {
		name           string
		authHeader     string
		expectedStatus int
		expectedUserID int
	}{
		{"No Authorization header", "", http.StatusOK, 0}, // Без токена, запрос проходит
		{"Invalid token format", "Bearer1234", http.StatusUnauthorized, 0},
		{"Malformed JWT", "Bearer invalid.token.string", http.StatusUnauthorized, 0},
		{"Expired token", "Bearer " + expiredToken, http.StatusUnauthorized, 0},
		{"Valid token", "Bearer " + validToken, http.StatusOK, 12345},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/", nil)
			req.Header.Set("Authorization", tt.authHeader)

			recorder := httptest.NewRecorder()
			mock := &mockHandler{}

			middleware.Middleware()(mock).ServeHTTP(recorder, req)

			assert.Equal(t, tt.expectedStatus, recorder.Code)

			if tt.expectedUserID != 0 {
				userID, _ := mock.ctx.Value(ctxkeys.UserIDKey).(int)
				assert.Equal(t, tt.expectedUserID, userID)
			}
		})
	}
}
