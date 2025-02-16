package middleware

import (
	"context"
	"encoding/json"
	"log"
	"merch-store-service/internal/api"
	"merch-store-service/internal/infra/jwtutils"
	"merch-store-service/pkg/ctxkeys"
	"net/http"
	"strings"
)

type AuthMiddleware struct {
	JWTManager *jwtutils.JWTManager
}

func NewAuthMiddleware(jwtManager *jwtutils.JWTManager) *AuthMiddleware {
	return &AuthMiddleware{JWTManager: jwtManager}
}

func writeError(w http.ResponseWriter, statusCode int, errMsg string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)

	if err := json.NewEncoder(w).Encode(api.ErrorResponse{
		Errors: &errMsg,
	}); err != nil {
		log.Println("Failed to encode JSON:", err)
	}
}

func (m *AuthMiddleware) Middleware() api.MiddlewareFunc {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")

			if authHeader == "" {
				next.ServeHTTP(w, r)
				return
			}

			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				writeError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			tokenString := tokenParts[1]

			claims, err := m.JWTManager.ValidateToken(tokenString)
			if err != nil {
				writeError(w, http.StatusUnauthorized, "Unauthorized")
				return
			}

			userID := claims.UserID

			ctx := context.WithValue(r.Context(), ctxkeys.UserIDKey, userID)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
