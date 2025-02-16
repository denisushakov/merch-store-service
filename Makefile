DB_URL="postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
MIGRATIONS_PATH="migrations"

.PHONY: migrate-up migrate-down migrate-force migrate-version

generateoapi:
	oapi-codegen --package=api \
		--generate types \
		-o internal/api/types.gen.go \
		internal/api/schema.yaml

	oapi-codegen --package=api \
		--generate chi-server \
		-o internal/api/server.gen.go \
		internal/api/schema.yaml

run:
	env CONFIG_PATH=configs/values_local.yaml go run cmd/merch-store/main.go

runtest:
	go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out | grep total

migrate-up:
	go run cmd/migrator/main.go -action=up -dburl=$(DB_URL) -migrationspath=$(MIGRATIONS_PATH)

migrate-down:
	go run cmd/migrator/main.go -action=down -dburl=$(DB_URL) -migrationspath=$(MIGRATIONS_PATH)

migrate-force:
	go run cmd/migrator/main.go -action=force -dburl=$(DB_URL) -migrationspath=$(MIGRATIONS_PATH) -version=$(VERSION)

migrate-version:
	go run cmd/migrator/main.go -action=version -dburl=$(DB_URL) -migrationspath=$(MIGRATIONS_PATH)