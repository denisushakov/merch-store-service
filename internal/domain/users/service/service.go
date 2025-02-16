package services

import (
	"context"
	"merch-store-service/internal/domain/repository"
	"merch-store-service/internal/infra/jwtutils"
	"time"

	"golang.org/x/crypto/bcrypt"
)

//go:generate go run github.com/vektra/mockery/v2@v2.48.0 --name=UserServiceAuth
type UserServiceAuth interface {
	PostApiAuth(ctx context.Context, username string, password string) (string, error)
}

type UserService struct {
	userRepo   *repository.Storage
	jwtManager *jwtutils.JWTManager
}

func NewUserService(userRepo *repository.Storage, jwtManager *jwtutils.JWTManager) *UserService {
	return &UserService{
		userRepo:   userRepo,
		jwtManager: jwtManager,
	}
}

func (s *UserService) PostApiAuth(ctx context.Context, username string, password string) (string, error) {
	user, err := s.userRepo.GetUserByUsername(ctx, username)
	if err != nil {
		hashedPassword, hashErr := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if hashErr != nil {
			return "", hashErr
		}

		userID, createErr := s.userRepo.CreateUser(ctx, username, string(hashedPassword))
		if createErr != nil {
			return "", createErr
		}

		return s.jwtManager.NewToken(userID, 24*time.Hour)
	}

	if compareErr := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); compareErr != nil {
		return "", repository.ErrUnauthorized
	}

	return s.jwtManager.NewToken(user.ID, 24*time.Hour)
}
