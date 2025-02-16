package repository

import (
	"context"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"merch-store-service/internal/api"
	"merch-store-service/internal/domain/models"
	"merch-store-service/internal/infra/config"

	"github.com/jackc/pgx/v5"
)

type Storage struct {
	db *pgxpool.Pool
}

func (s *Storage) DB() *pgxpool.Pool {
	return s.db
}

func New(cfg *config.Config) (*Storage, error) {
	const op = "domain.repository.New"

	connConfig, err := pgxpool.ParseConfig(fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		cfg.DBUser, cfg.DBPassword, cfg.DBHost, cfg.DBPort, cfg.DBName))
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	conn, err := pgxpool.NewWithConfig(context.Background(), connConfig)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	return &Storage{db: conn}, nil
}

func (s *Storage) GetUserCoins(ctx context.Context, userID int) (int, error) {
	const op = "domain.repository.GetUserCoins"

	var coins int
	err := s.db.QueryRow(ctx, "SELECT coins FROM users WHERE id=$1", userID).Scan(&coins)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			return 0, fmt.Errorf("%s: user with ID %d not found", op, userID)
		}
		return 0, fmt.Errorf("%s: failed to get user coins: %w", op, err)
	}

	return coins, nil
}

func (s *Storage) BuyItem(ctx context.Context, userID int, item string) error {
	const op = "domain.repository.BuyItem"

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: failed to begin transaction: %w", op, err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback(ctx)
		}
	}()

	var price int
	err = tx.QueryRow(ctx, "SELECT price FROM products WHERE name=$1", item).Scan(&price)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return fmt.Errorf("%s: item not found", op)
		}
		return fmt.Errorf("%s: failed to get item price: %w", op, err)
	}

	var userCoins int
	err = tx.QueryRow(ctx, "UPDATE users SET coins = coins - $1 WHERE id = $2 AND coins >= $1 RETURNING coins", price, userID).Scan(&userCoins)
	if err != nil {
		return fmt.Errorf("%s: failed to update user coins: %w", op, err)
	}

	_, err = tx.Exec(ctx, `
        INSERT INTO inventory (user_id, item_name)
        VALUES ($1, $2)
        ON CONFLICT (user_id, item_name)
        DO UPDATE SET quantity = inventory.quantity + 1`, userID, item)
	if err != nil {
		return fmt.Errorf("%s: failed to add item to inventory: %w", op, err)
	}

	_, err = tx.Exec(ctx, "INSERT INTO transactions (from_user_id, to_user_id, amount) VALUES ($1, $2, $3)", userID, userID, price)
	if err != nil {
		return fmt.Errorf("%s: failed to insert transaction: %w", op, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: failed to commit transaction: %w", op, err)
	}

	return nil
}

//nolint:gocyclo
func (s *Storage) GetUserCoinHistory(ctx context.Context, userID int) (*api.InfoResponse, error) {
	const op = "domain.repository.GetUserCoinHistory"

	coinHistory := &api.InfoResponse{
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
	}

	rows, err := s.db.Query(ctx, "SELECT amount, from_user_id FROM transactions WHERE to_user_id=$1", userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	coinHistory.CoinHistory.Received = &[]struct {
		Amount   *int    `json:"amount,omitempty"`
		FromUser *string `json:"fromUser,omitempty"`
	}{}

	for rows.Next() {
		var amount int
		var fromUserID int
		if err := rows.Scan(&amount, &fromUserID); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		fromUser, err := s.GetUserByID(ctx, fromUserID)
		if errors.Is(err, ErrUserNotFound) {
			log.Printf("%s: user not found, skipping", op)
			continue
		} else if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		*coinHistory.CoinHistory.Received = append(*coinHistory.CoinHistory.Received, struct {
			Amount   *int    `json:"amount,omitempty"`
			FromUser *string `json:"fromUser,omitempty"`
		}{
			Amount:   &amount,
			FromUser: &fromUser,
		})
	}

	rows.Close()

	rows, err = s.db.Query(ctx, "SELECT amount, to_user_id FROM transactions WHERE from_user_id=$1", userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	coinHistory.CoinHistory.Sent = &[]struct {
		Amount *int    `json:"amount,omitempty"`
		ToUser *string `json:"toUser,omitempty"`
	}{}

	for rows.Next() {
		var amount int
		var toUserID int
		if err := rows.Scan(&amount, &toUserID); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		toUser, err := s.GetUserByID(ctx, toUserID)
		if errors.Is(err, ErrUserNotFound) {
			log.Printf("%s: user not found, skipping", op)
			continue
		} else if err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}
		*coinHistory.CoinHistory.Sent = append(*coinHistory.CoinHistory.Sent, struct {
			Amount *int    `json:"amount,omitempty"`
			ToUser *string `json:"toUser,omitempty"`
		}{
			Amount: &amount,
			ToUser: &toUser,
		})
	}

	return coinHistory, nil
}

func (s *Storage) SendCoins(ctx context.Context, fromUserID, toUserID, amount int) error {
	const op = "domain.repository.SendCoins"

	tx, err := s.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("%s: failed to begin transaction: %w", op, err)
	}
	defer func() {
		if err != nil {
			err := tx.Rollback(ctx)
			if err != nil {
				log.Printf("%s: rollback failed: %v", op, err)
			}
		}
	}()

	var senderCoins int
	err = tx.QueryRow(ctx, "SELECT coins FROM users WHERE id=$1 FOR UPDATE", fromUserID).Scan(&senderCoins)
	if err != nil {
		return fmt.Errorf("%s: failed to get user's coins: %w", op, err)
	}

	if senderCoins < amount {
		return fmt.Errorf("%s: insufficient funds", op)
	}

	_, err = tx.Exec(ctx, "UPDATE users SET coins = coins - $1 WHERE id = $2", amount, fromUserID)
	if err != nil {
		return fmt.Errorf("%s: failed to update sender's coins: %w", op, err)
	}

	_, err = tx.Exec(ctx, "UPDATE users SET coins = coins + $1 WHERE id = $2", amount, toUserID)
	if err != nil {
		return fmt.Errorf("%s: failed to update recipient's coins: %w", op, err)
	}

	_, err = tx.Exec(ctx, "INSERT INTO transactions (from_user_id, to_user_id, amount) VALUES ($1, $2, $3)", fromUserID, toUserID, amount)
	if err != nil {
		return fmt.Errorf("%s: failed to insert transaction: %w", op, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("%s: failed to commit transaction: %w", op, err)
	}

	return nil
}

func (s *Storage) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	const op = "domain.repository.GetUserByUsername"

	var user models.User
	err := s.db.QueryRow(ctx,
		"SELECT id, username, password_hash FROM users WHERE username=$1",
		username,
	).Scan(&user.ID, &user.Username, &user.PasswordHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}

		return nil, fmt.Errorf("%s: execute statement: %w", op, err)
	}
	return &user, nil
}

func (s *Storage) GetUserByID(ctx context.Context, userID int) (string, error) {
	const op = "domain.repository.GetUserByID"

	var username string
	err := s.db.QueryRow(ctx, "SELECT username FROM users WHERE id=$1", userID).Scan(&username)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", fmt.Errorf("%s: %w", op, ErrUserNotFound)
		}
		return "", fmt.Errorf("%s: %w", op, err)
	}
	return username, nil
}

func (s *Storage) CreateUser(ctx context.Context, username, passwordHash string) (int, error) {
	const op = "domain.repository.CreateUser"

	var userID int
	err := s.db.QueryRow(ctx, "INSERT INTO users(username, password_hash) VALUES  ($1, $2) RETURNING id", username, passwordHash).Scan(&userID)
	if err != nil {
		return 0, fmt.Errorf("%s: %w", op, err)
	}

	return userID, nil
}

func (s *Storage) GetUserInventory(ctx context.Context, userID int) (*api.InfoResponse, error) {
	const op = "domain.repository.GetUserInventory"

	coinHistory := &api.InfoResponse{
		Inventory: &[]struct {
			Quantity *int    `json:"quantity,omitempty"`
			Type     *string `json:"type,omitempty"`
		}{},
	}

	rows, err := s.db.Query(ctx, "SELECT item_name, quantity FROM inventory WHERE user_id=$1", userID)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	defer rows.Close()

	for rows.Next() {
		var item struct {
			ItemName string
			Quantity int
		}
		if err := rows.Scan(&item.ItemName, &item.Quantity); err != nil {
			return nil, fmt.Errorf("%s: %w", op, err)
		}

		*coinHistory.Inventory = append(*coinHistory.Inventory, struct {
			Quantity *int    `json:"quantity,omitempty"`
			Type     *string `json:"type,omitempty"`
		}{
			Quantity: &item.Quantity,
			Type:     &item.ItemName,
		})
	}

	return coinHistory, nil
}
