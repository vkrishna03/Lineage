# Getting Started

## Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Kafka (or Aiven Kafka)
- Make

## Setup

### 1. Clone and install tools

```bash
cd apps/api
make tools  # Installs sqlc, migrate, swag
```

### 2. Configure environment

```bash
cp .env.example .env
# Edit .env with your database and Kafka settings
```

Required environment variables:
- `DATABASE_URL`: Postgres connection string
- `KAFKA_BROKERS`: Comma-separated broker addresses
- `KAFKA_TOPIC`: Topic name (default: `lineage.events`)

For Aiven Kafka:
- `KAFKA_SASL_ENABLED=true`
- `KAFKA_SASL_USERNAME`: Your Aiven username
- `KAFKA_SASL_PASSWORD`: Your Aiven password
- `KAFKA_CA_PATH`: Path to Aiven CA certificate

### 3. Run migrations

```bash
make migrate-up
```

### 4. Start services

In separate terminals:

```bash
# Terminal 1: API server
make run-api

# Terminal 2: Kafka consumer
make run-consumer
```

## Verify

### Health check

```bash
curl http://localhost:8080/health
```

### View API docs

Open http://localhost:8080/swagger/index.html

## Quick Test

### 1. Create a scope

```bash
curl -X POST http://localhost:8080/api/v1/scopes \
  -H "Content-Type: application/json" \
  -d '{"project": "test-project"}'
```

### 2. Create an actor

```bash
curl -X POST http://localhost:8080/api/v1/actors \
  -H "Content-Type: application/json" \
  -d '{"type": "llm", "name": "Claude"}'
```

### 3. Create an event type

```bash
curl -X POST http://localhost:8080/api/v1/event-types \
  -H "Content-Type: application/json" \
  -d '{"name": "decision", "version": "1.0"}'
```

### 4. Submit an event

```bash
curl -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -d '{
    "scope_id": "<scope-id>",
    "actor_id": "<actor-id>",
    "event_type_id": "<event-type-id>",
    "intent": "decision",
    "payload": {"action": "approved", "confidence": 0.95}
  }'
```

## Development

### Generate code

```bash
make generate  # Runs sqlc and swagger generation
```

### Run tests

```bash
make test
```

### Create migration

```bash
make migrate-create name=add_new_table
```
