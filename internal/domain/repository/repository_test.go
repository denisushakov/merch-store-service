package repository_test

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"log"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/ory/dockertest/v3"
	"github.com/stretchr/testify/assert"
	"merch-store-service/internal/domain/repository"
	"merch-store-service/internal/infra/config"
)

// SetupTestDatabase создает контейнер PostgreSQL с использованием Dockertest и выполняет SQL-запросы для создания таблиц
func SetupTestDatabase() (*pgxpool.Pool, string, func()) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to Docker: %s", err)
	}

	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "15",
		Env: []string{
			"POSTGRES_USER=postgres",
			"POSTGRES_PASSWORD=postgres",
			"POSTGRES_DB=avito_db",
		},
	})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	dbPort := resource.GetPort("5432/tcp")
	dsn := fmt.Sprintf("postgresql://postgres:postgres@localhost:%s/postgres?sslmode=disable", dbPort)

	var db *pgxpool.Pool
	err = pool.Retry(func() error {
		ctx := context.Background()
		db, err = pgxpool.New(ctx, dsn)
		if err != nil {
			return err
		}
		return db.Ping(ctx)
	})

	if err != nil {
		log.Fatalf("Could not connect to database after retries: %s", err)
	}

	log.Println("PostgreSQL is up and running")

	teardown := func() {
		log.Println("Cleaning up database resources...")
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}
	}

	return db, dbPort, teardown
}

func createTables(db *pgxpool.Pool) error {
	_, err := db.Exec(context.Background(), `
		CREATE TABLE IF NOT EXISTS users
		(
			id SERIAL PRIMARY KEY,
			username VARCHAR(255) NOT NULL UNIQUE,
			password_hash VARCHAR(255) NOT NULL,
			coins INT DEFAULT 1000,
			created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP
		);

		CREATE TABLE IF NOT EXISTS transactions
		(
			id SERIAL PRIMARY KEY,
			from_user_id INT NOT NULL,
			to_user_id INT NOT NULL,
			amount INT NOT NULL,
			created_at TIMESTAMP WITHOUT TIME ZONE DEFAULT CURRENT_TIMESTAMP,
			FOREIGN KEY (from_user_id) REFERENCES users(id) ON DELETE CASCADE,
			FOREIGN KEY (to_user_id) REFERENCES users(id) ON DELETE CASCADE
		);

		CREATE TABLE IF NOT EXISTS inventory
		(
			id SERIAL PRIMARY KEY,
			user_id INT NOT NULL,
			item_name VARCHAR(255) NOT NULL,
			quantity INT NOT NULL DEFAULT 1,
			FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
			CONSTRAINT unique_inventory UNIQUE(user_id, item_name)
		);

		CREATE TABLE IF NOT EXISTS products
		(
			id SERIAL PRIMARY KEY,
			name VARCHAR(255) NOT NULL UNIQUE,
			price INT NOT NULL,
		    CONSTRAINT unique_name UNIQUE(name)
		);

		INSERT INTO products (name, price) VALUES
			('t-shirt', 80),
			('cup', 20),
			('book', 50),
			('pen', 10),
			('powerbank', 200),
			('hoody', 300),
			('umbrella', 200),
			('socks', 10),
			('wallet', 50),
			('pink-hoody', 500)
		ON CONFLICT (name) DO NOTHING;
	`)
	return err
}

func cleanDatabase(db *pgxpool.Pool) error {
	_, err := db.Exec(context.Background(), `
		TRUNCATE TABLE users RESTART IDENTITY CASCADE;
		TRUNCATE TABLE transactions RESTART IDENTITY CASCADE;
		TRUNCATE TABLE inventory RESTART IDENTITY CASCADE;
	`)
	return err
}

func TestRepository(t *testing.T) {
	db, dbPort, teardown := SetupTestDatabase()
	defer teardown()

	err := createTables(db)
	assert.NoError(t, err, "Table creation should execute without errors")

	err = cleanDatabase(db)
	assert.NoError(t, err, "Database cleanup should execute without errors")

	cfg := &config.Config{
		DBUser:     "postgres",
		DBPassword: "postgres",
		DBHost:     "localhost",
		DBPort:     dbPort,
		DBName:     "postgres",
	}
	storage, err := repository.New(cfg)
	assert.NoError(t, err, "Database connection should be successful")

	ctx := context.Background()

	var count int
	err = db.QueryRow(ctx, "SELECT COUNT(*) FROM products").Scan(&count)
	assert.NoError(t, err, "Error querying product count")
	assert.Equal(t, 10, count, "The store should have exactly 10 products")

	var price int
	err = db.QueryRow(ctx, "SELECT price FROM products WHERE name = 'pen'").Scan(&price)
	assert.NoError(t, err, "Error retrieving price for 'pen'")
	assert.Equal(t, 10, price, "The price of 'pen' should be fixed at 10 coins")

	username := uuid.New().String()
	userID, err := storage.CreateUser(ctx, username, "password_hash")
	assert.NoError(t, err, "CreateUser should work correctly")

	coins, err := storage.GetUserCoins(ctx, userID)
	assert.NoError(t, err)
	assert.Equal(t, 1000, coins, "Initial balance should be 1000 coins")

	err = storage.BuyItem(ctx, userID, "pen")
	assert.NoError(t, err, "BuyItem should work correctly")

	newCoins, _ := storage.GetUserCoins(ctx, userID)
	assert.Equal(t, 990, newCoins, "After purchasing 'pen', balance should be 990 coins")

	err = db.QueryRow(ctx, "SELECT COUNT(*) FROM inventory WHERE user_id = $1 AND item_name = 'pen'", userID).Scan(&count)
	assert.NoError(t, err)
	assert.Greater(t, count, 0, "User should have 'pen' in their inventory")

	user2ID, _ := storage.CreateUser(ctx, uuid.New().String(), "password_hash")
	err = storage.SendCoins(ctx, userID, user2ID, 100)
	assert.NoError(t, err)

	user1Balance, _ := storage.GetUserCoins(ctx, userID)
	user2Balance, _ := storage.GetUserCoins(ctx, user2ID)
	assert.Equal(t, 890, user1Balance, "After transferring 100 coins, the first user should have 890 coins left")
	assert.Equal(t, 1100, user2Balance, "After receiving 100 coins, the second user should have 1100 coins")

	coinHistory, err := storage.GetUserCoinHistory(ctx, userID)
	assert.NoError(t, err)
	assert.NotEmpty(t, coinHistory.CoinHistory.Sent, "Sent transaction history should not be empty")

	inventory, err := storage.GetUserInventory(ctx, userID)
	assert.NoError(t, err, "GetUserInventory should work correctly")
	assert.NotEmpty(t, inventory.Inventory, "Inventory should not be empty")

	log.Println("All tests passed successfully!")
}
