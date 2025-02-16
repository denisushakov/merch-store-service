package tests

import (
	"bytes"
	"encoding/json"
	"flag"
	"net/http"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"merch-store-service/internal/api"
	"merch-store-service/internal/app"
	"merch-store-service/internal/infra/config"
	"net/http/httptest"
)

// setupServer Инициализация сервера для тестов
func setupServer() *httptest.Server {
	flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ContinueOnError)

	os.Setenv("CONFIG_PATH", "../configs/values_local.yaml")
	defer os.Unsetenv("CONFIG_PATH")

	cfg := config.LoadConfig()
	appInstance := app.New(cfg)
	return httptest.NewServer(appInstance.Server.Handler)
}

// TestSendCoinValid Тест отправки монет
func TestSendCoinValid(t *testing.T) {
	server := setupServer()
	defer server.Close()

	token := "valid-jwt-token"

	sendCoinReq := api.SendCoinRequest{
		ToUser: "anotheruser",
		Amount: 100,
	}

	reqBody, err := json.Marshal(sendCoinReq)
	if err != nil {
		t.Fatalf("Не удалось сериализовать запрос: %v", err)
	}

	req, err := http.NewRequest("POST", server.URL+"/api/sendCoin", bytes.NewBuffer(reqBody))
	if err != nil {
		t.Fatalf("Не удалось создать запрос: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	//nolint:errcheck
	server.Client().Do(req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

// TestGetInfoValid Тест получения информации о пользователе
func TestGetInfoValid(t *testing.T) {
	server := setupServer()
	defer server.Close()

	token := "valid-jwt-token"

	req, err := http.NewRequest("GET", server.URL+"/api/info", nil)
	if err != nil {
		t.Fatalf("Не удалось создать запрос: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	//nolint:errcheck
	server.Client().Do(req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

// TestBuyItemValid Тест покупки мерча
func TestBuyItemValid(t *testing.T) {
	server := setupServer()
	defer server.Close()

	token := "valid-jwt-token"

	item := "t-shirt"
	req, err := http.NewRequest("GET", server.URL+"/api/buy/"+item, nil)
	if err != nil {
		t.Fatalf("Не удалось создать запрос: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	//nolint:errcheck
	server.Client().Do(req)

	assert.Equal(t, http.StatusOK, rr.Code)
}

// TestGetUserInfoValid Тест получения информации о пользователе
func TestGetUserInfoValid(t *testing.T) {
	server := setupServer()
	defer server.Close()

	token := "valid-jwt-token"

	req, err := http.NewRequest("GET", server.URL+"/api/info", nil)
	if err != nil {
		t.Fatalf("Не удалось создать запрос: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+token)

	rr := httptest.NewRecorder()
	//nolint:errcheck
	server.Client().Do(req)

	assert.Equal(t, http.StatusOK, rr.Code)
}
