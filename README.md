# Lineage

Epistemic Transparency & Event Lineage System — an append-only event store with hash-chaining for tracking AI decisions.

## Features

- **Hash-chained events**: Each event links to the previous via SHA-256 hash (RFC 8785 canonical JSON)
- **Epistemic intents**: Track exploration, suggestion, assertion, decision, and execution
- **Lineage tracking**: Parent/child relationships between events
- **Corrections**: Supersede, amend, or retract previous events
- **Async processing**: Kafka-based event ingestion

## Quick Start

```bash
cd apps/api
cp .env.example .env  # Configure database and Kafka
make tools            # Install sqlc, migrate, swag
make migrate-up       # Run database migrations
make run-api          # Start API server
make run-consumer     # Start Kafka consumer (separate terminal)
```

Visit http://localhost:8080/swagger/index.html for API docs.

## Architecture

```
Client → API (Gin) → Kafka → Consumer → Postgres
```

See [docs/architecture.md](./docs/architecture.md) for details.

## Documentation

- [Architecture](./docs/architecture.md) - System design
- [Getting Started](./docs/getting-started.md) - Setup guide
- [API Reference](./docs/api.md) - REST endpoints

## Tech Stack

- **API**: Go, Gin, SQLC
- **Database**: PostgreSQL
- **Queue**: Kafka (Aiven)
- **Docs**: Swagger/OpenAPI

## License

[Elastic License 2.0](./LICENSE)
