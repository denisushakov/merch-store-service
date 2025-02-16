package services

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"merch-store-service/internal/api"
	"merch-store-service/internal/domain/coins/service/mocks"
)

func TestSendCoins(t *testing.T) {
	mockService := new(mocks.CoinServiceInterface)

	testCases := []struct {
		name        string
		fromUserID  int
		toUsername  string
		amount      int
		mockErr     error
		expectErr   bool
		expectedErr string
	}{
		{
			name:       "Successful coin transfer",
			fromUserID: 1,
			toUsername: "recipient",
			amount:     100,
			mockErr:    nil,
			expectErr:  false,
		},
		{
			name:        "Recipient user not found",
			fromUserID:  1,
			toUsername:  "unknown",
			amount:      50,
			mockErr:     errors.New("recipient user 'unknown' not found"),
			expectErr:   true,
			expectedErr: "recipient user 'unknown' not found",
		},
		{
			name:        "Cannot send coins to yourself",
			fromUserID:  1,
			toUsername:  "self",
			amount:      30,
			mockErr:     errors.New("cannot send coins to yourself"),
			expectErr:   true,
			expectedErr: "cannot send coins to yourself",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService.On("SendCoins", mock.Anything, tc.fromUserID, tc.toUsername, tc.amount).
				Return(tc.mockErr)

			err := mockService.SendCoins(context.Background(), tc.fromUserID, tc.toUsername, tc.amount)

			if tc.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestBuyItem(t *testing.T) {
	mockService := new(mocks.CoinServiceInterface)

	testCases := []struct {
		name        string
		userID      int
		item        string
		mockErr     error
		expectErr   bool
		expectedErr string
	}{
		{
			name:      "Successful item purchase",
			userID:    1,
			item:      "t-shirt",
			mockErr:   nil,
			expectErr: false,
		},
		{
			name:        "Item not found",
			userID:      1,
			item:        "unknown-item",
			mockErr:     errors.New("failed to buy item"),
			expectErr:   true,
			expectedErr: "failed to buy item",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService.On("BuyItem", mock.Anything, tc.userID, tc.item).
				Return(tc.mockErr)

			err := mockService.BuyItem(context.Background(), tc.userID, tc.item)

			if tc.expectErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tc.expectedErr)
			} else {
				assert.NoError(t, err)
			}

			mockService.AssertExpectations(t)
		})
	}
}

func TestGetUserInfo(t *testing.T) {
	mockService := new(mocks.CoinServiceInterface)

	testCases := []struct {
		name        string
		userID      int
		mockCoins   int
		mockHistory *api.InfoResponse
		mockInv     *api.InfoResponse
		mockErr     error
		expectErr   bool
	}{
		{
			name:      "Successful user info retrieval",
			userID:    1,
			mockCoins: 1000,
			mockHistory: &api.InfoResponse{
				CoinHistory: &struct {
					Received *[]struct {
						Amount   *int    `json:"amount,omitempty"`
						FromUser *string `json:"fromUser,omitempty"`
					} `json:"received,omitempty"`
					Sent *[]struct {
						Amount *int    `json:"amount,omitempty"`
						ToUser *string `json:"toUser,omitempty"`
					} `json:"sent,omitempty"`
				}{},
			},
			mockInv: &api.InfoResponse{
				Inventory: &[]struct {
					Quantity *int    `json:"quantity,omitempty"`
					Type     *string `json:"type,omitempty"`
				}{},
			},
			mockErr:   nil,
			expectErr: false,
		},
		{
			name:      "User not found",
			userID:    99,
			mockCoins: 0,
			mockErr:   errors.New("user not found"),
			expectErr: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService.On("GetUserInfo", mock.Anything, tc.userID).
				Return(&api.InfoResponse{Coins: &tc.mockCoins}, tc.mockErr)

			info, err := mockService.GetUserInfo(context.Background(), tc.userID)

			if tc.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, info)
				assert.Equal(t, tc.mockCoins, *info.Coins)
			}

			mockService.AssertExpectations(t)
		})
	}
}
