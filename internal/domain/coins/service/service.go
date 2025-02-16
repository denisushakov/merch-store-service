package services

import (
	"context"
	"errors"
	"fmt"
	"merch-store-service/internal/api"
	"merch-store-service/internal/domain/repository"
)

//go:generate go run github.com/vektra/mockery/v2@v2.48.0 --name=CoinServiceInterface
type CoinServiceInterface interface {
	SendCoins(ctx context.Context, fromUserID int, toUser string, amount int) error
	BuyItem(ctx context.Context, userID int, item string) error
	GetUserInfo(ctx context.Context, userID int) (*api.InfoResponse, error)
}

type CoinService struct {
	storage *repository.Storage
}

func NewCoinService(storage *repository.Storage) *CoinService {
	return &CoinService{
		storage: storage,
	}
}

func (s *CoinService) SendCoins(ctx context.Context, fromUserID int, toUser string, amount int) error {

	user, err := s.storage.GetUserByUsername(ctx, toUser)
	if err != nil {
		if errors.Is(err, repository.ErrUserNotFound) {
			return fmt.Errorf("recipient user '%s' not found", toUser)
		}
		return fmt.Errorf("failed to get recipient: %w", err)
	}

	if fromUserID == user.ID {
		return fmt.Errorf("cannot send coins to yourself")
	}

	return s.storage.SendCoins(ctx, fromUserID, user.ID, amount)
}

func (s *CoinService) BuyItem(ctx context.Context, userID int, item string) error {
	return s.storage.BuyItem(ctx, userID, item)
}

func (s *CoinService) GetUserInfo(ctx context.Context, userID int) (*api.InfoResponse, error) {
	coins, err := s.storage.GetUserCoins(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user coins: %w", err)
	}

	coinHistory, err := s.storage.GetUserCoinHistory(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user coin history: %w", err)
	}

	inventoryResponse, err := s.storage.GetUserInventory(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user inventory: %w", err)
	}

	return &api.InfoResponse{
		Coins:       &coins,
		CoinHistory: coinHistory.CoinHistory,
		Inventory:   inventoryResponse.Inventory,
	}, nil
}
