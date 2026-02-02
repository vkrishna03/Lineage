.PHONY: run-api run-consumer build sqlc swagger generate migrate-up migrate-down migrate-create test lint tidy clean help

# Load .env file if it exists
ifneq (,$(wildcard .env))
    include .env
    export
endif

# Default database URL for migrations
DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/lineage-db?sslmode=disable

# Kafka settings
KAFKA_BROKERS ?= localhost:9092
KAFKA_TOPIC ?= lineage.events

# Run the API server
run-api:
	go run ./cmd/api

# Run the Kafka consumer
run-consumer:
	go run ./cmd/consumer

# Build both binaries
build:
	go build -o bin/api ./cmd/api
	go build -o bin/consumer ./cmd/consumer

# Generate SQLC code
sqlc:
	sqlc generate

# Generate Swagger documentation
swagger:
	swag init -g cmd/api/main.go -o docs

# Generate all code
generate: sqlc swagger

# Run database migrations up
migrate-up:
	migrate -path internal/app/db/migrations -database "$(DATABASE_URL)" up

# Run database migrations down
migrate-down:
	migrate -path internal/app/db/migrations -database "$(DATABASE_URL)" down

# Run a specific number of migrations down
migrate-down-one:
	migrate -path internal/app/db/migrations -database "$(DATABASE_URL)" down 1

# Create a new migration
# Usage: make migrate-create name=add_new_table
migrate-create:
	migrate create -ext sql -dir internal/app/db/migrations -seq $(name)

# Force migration version (use with caution)
# Usage: make migrate-force version=1
migrate-force:
	migrate -path internal/app/db/migrations -database "$(DATABASE_URL)" force $(version)

# Run tests
test:
	go test -v ./...

# Run linter
lint:
	golangci-lint run

# Tidy dependencies
tidy:
	go mod tidy

# Clean build artifacts
clean:
	rm -rf bin/

# Install dev tools
tools:
	go install github.com/sqlc-dev/sqlc/cmd/sqlc@latest
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	go install github.com/swaggo/swag/cmd/swag@latest

# Show help
help:
	@echo "Available targets:"
	@echo "  run-api        - Run the API server"
	@echo "  run-consumer   - Run the Kafka consumer"
	@echo "  build          - Build both binaries"
	@echo "  sqlc           - Generate SQLC code"
	@echo "  swagger        - Generate Swagger documentation"
	@echo "  generate       - Generate all code (SQLC, Swagger)"
	@echo "  migrate-up     - Run all pending migrations"
	@echo "  migrate-down   - Rollback all migrations"
	@echo "  migrate-create - Create a new migration (name=xxx)"
	@echo "  test           - Run tests"
	@echo "  lint           - Run linter"
	@echo "  tidy           - Tidy Go modules"
	@echo "  tools          - Install dev tools (sqlc, migrate)"
	@echo "  clean          - Clean build artifacts"
