.PHONY: run-api run-consumer build sqlc migrate-up migrate-down migrate-create kafka-topics test lint

# Load .env file if it exists
ifneq (,$(wildcard .env))
    include .env
    export
endif

# Default database URL for migrations
DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/lineage?sslmode=disable

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

# Run database migrations up
migrate-up:
	migrate -path internal/db/migrations -database "$(DATABASE_URL)" up

# Run database migrations down
migrate-down:
	migrate -path internal/db/migrations -database "$(DATABASE_URL)" down

# Run a specific number of migrations down
migrate-down-one:
	migrate -path internal/db/migrations -database "$(DATABASE_URL)" down 1

# Create a new migration
# Usage: make migrate-create name=add_new_table
migrate-create:
	migrate create -ext sql -dir internal/db/migrations -seq $(name)

# Force migration version (use with caution)
# Usage: make migrate-force version=1
migrate-force:
	migrate -path internal/db/migrations -database "$(DATABASE_URL)" force $(version)

# Create Kafka topics
kafka-topics:
	kafka-topics.sh --create --topic $(KAFKA_TOPIC) --bootstrap-server $(KAFKA_BROKERS) --partitions 16 --replication-factor 1 --if-not-exists || \
	docker exec -it kafka kafka-topics.sh --create --topic $(KAFKA_TOPIC) --bootstrap-server localhost:9092 --partitions 16 --replication-factor 1 --if-not-exists || \
	echo "Note: Create topic manually if Kafka CLI not available"

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

# Show help
help:
	@echo "Available targets:"
	@echo "  run-api        - Run the API server"
	@echo "  run-consumer   - Run the Kafka consumer"
	@echo "  build          - Build both binaries"
	@echo "  sqlc           - Generate SQLC code"
	@echo "  migrate-up     - Run all pending migrations"
	@echo "  migrate-down   - Rollback all migrations"
	@echo "  migrate-create - Create a new migration (name=xxx)"
	@echo "  kafka-topics   - Create Kafka topics"
	@echo "  test           - Run tests"
	@echo "  lint           - Run linter"
	@echo "  tidy           - Tidy Go modules"
	@echo "  clean          - Clean build artifacts"
