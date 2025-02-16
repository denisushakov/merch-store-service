package app

import (
	"encoding/json"
	"fmt"
	"log"
	"merch-store-service/internal/api"
	coinService "merch-store-service/internal/domain/coins/service"
	userService "merch-store-service/internal/domain/users/service"
	"merch-store-service/pkg/ctxkeys"
	"net/http"
)

type Server struct {
	UserService *userService.UserService
	CoinService *coinService.CoinService
}

// PostApiAuth Аутентификация и получение JWT-токена. При первой аутентификации пользователь создается автоматически.
// (POST /api/auth)
func (s *Server) PostApiAuth(w http.ResponseWriter, r *http.Request) {
	if r.Body == nil {
		http.Error(w, "Request body is empty", http.StatusBadRequest)
		return
	}

	var req api.AuthRequest
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(&req); err != nil {
		http.Error(w, "Invalid JSON format", http.StatusBadRequest)
		return
	}

	if req.Username == "" || req.Password == "" {
		http.Error(w, "Username and password are required", http.StatusBadRequest)
		return
	}

	token, err := s.UserService.PostApiAuth(r.Context(), req.Username, req.Password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	resp := api.AuthResponse{Token: &token}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Failed to encode response: %v", err)
	}
}

// GetApiBuyItem Купить предмет за монеты.
// (GET /api/buy/{item})
func (s *Server) GetApiBuyItem(w http.ResponseWriter, r *http.Request, item string) {
	userID, ok := r.Context().Value(ctxkeys.UserIDKey).(int)
	if !ok {
		http.Error(w, "invalid user ID", http.StatusUnauthorized)
		return
	}

	err := s.CoinService.BuyItem(r.Context(), userID, item)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.WriteHeader(http.StatusOK)
}

// GetApiInfo Получить информацию о монетах, инвентаре и истории транзакций.
// (GET /api/info)
func (s *Server) GetApiInfo(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(ctxkeys.UserIDKey).(int)
	if !ok {
		http.Error(w, `{"errors": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	userInfo, err := s.CoinService.GetUserInfo(r.Context(), userID)
	if err != nil {
		http.Error(w, fmt.Sprintf(`{"errors": "%s"}`, err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(userInfo); err != nil {
		http.Error(w, fmt.Sprintf(`{"errors": "%s"}`, err.Error()), http.StatusInternalServerError)
	}
}

// PostApiSendCoin Отправить монеты другому пользователю.
// (POST /api/sendCoin)
func (s *Server) PostApiSendCoin(w http.ResponseWriter, r *http.Request) {
	var req api.SendCoinRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	userID, ok := r.Context().Value(ctxkeys.UserIDKey).(int)
	if !ok {
		http.Error(w, "invalid user ID", http.StatusUnauthorized)
		return
	}

	err := s.CoinService.SendCoins(r.Context(), userID, req.ToUser, req.Amount)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
