package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"merch-store-service/internal/domain/users/service/mocks"
)

func TestPostApiAuth(t *testing.T) {
	mockService := new(mocks.UserServiceAuth)

	// Определяем тест-кейсы
	testCases := []struct {
		name        string
		username    string
		password    string
		mockReturn  string
		mockError   error
		expectError bool
	}{
		{
			name:        "Successful authentication",
			username:    "testuser",
			password:    "password123",
			mockReturn:  "valid-token",
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "User not found - should create new user",
			username:    "newuser",
			password:    "newpassword",
			mockReturn:  "new-user-token",
			mockError:   nil,
			expectError: false,
		},
		{
			name:        "Incorrect password",
			username:    "testuser",
			password:    "wrongpassword",
			mockReturn:  "",
			mockError:   errors.New("unauthorized"),
			expectError: true,
		},
		{
			name:        "Database error",
			username:    "erroruser",
			password:    "password",
			mockReturn:  "",
			mockError:   errors.New("database error"),
			expectError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			mockService.On("PostApiAuth", mock.Anything, tc.username, tc.password).
				Return(tc.mockReturn, tc.mockError)

			token, err := mockService.PostApiAuth(context.Background(), tc.username, tc.password)

			if tc.expectError {
				assert.Error(t, err)
				assert.Empty(t, token)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.mockReturn, token)
			}

			mockService.AssertExpectations(t)
		})
	}
}
