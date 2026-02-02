# Lineage

Epistemic Transparency & Event Lineage System — an append-only event store with hash-chaining for tracking AI decisions.

## Features

- **Hash-chained events**: Each event links to the previous via SHA-256 hash (RFC 8785 canonical JSON)
- **Epistemic intents**: Track exploration, suggestion, assertion, decision, and execution
- **Lineage tracking**: Parent/child relationships between events
- **Corrections**: Supersede, amend, or retract previous events
- **Async processing**: Kafka-based event ingestion

## Quick Start with SDK

**Python:**
```bash
pip install lineage-sdk
```

```python
import lineage

lineage.init(project="my-app", actor_name="my-service", actor_type="service")
lineage.emit("data_ingestion", "assertion", {"data": "loaded"}, confidence=0.99)
```

**TypeScript:**
```bash
npm install lineage-sdk
```

```typescript
import * as lineage from 'lineage-sdk';

await lineage.init({ project: 'my-app', actorName: 'my-service', actorType: 'service' });
await lineage.emit('data_ingestion', 'assertion', { data: 'loaded' }, { confidence: 0.99 });
```

See [SDK Reference](./docs/sdk.md) for full documentation.

## Running the Server

```bash
cd apps/api
cp .env.example .env  # Configure database and Kafka
make tools            # Install sqlc, migrate, swag
make migrate-up       # Run database migrations
make run-api          # Start API server
make run-consumer     # Start Kafka consumer (separate terminal)
```

Visit http://localhost:8080/swagger/index.html for API docs.

## SDKs

| Language | Package | Install |
|----------|---------|---------|
| Python | [lineage-sdk](./sdks/python) | `pip install lineage-sdk` |
| TypeScript | [lineage-sdk](./sdks/typescript) | `npm install lineage-sdk` |
| curl | [Examples](./docs/sdk.md#curl-examples) | Direct HTTP |

## Examples

### Python
| Example | Description |
|---------|-------------|
| [pricing-decision-py](./examples/pricing-decision-py) | Basic pricing workflow with multiple actors |
| [langgraph-support-py](./examples/langgraph-support-py) | LangGraph customer support agent with human escalation |
| [rag-legal-py](./examples/rag-legal-py) | RAG pipeline for legal documents with provenance |
| [fastapi-service-py](./examples/fastapi-service-py) | FastAPI service with approval workflow |

### TypeScript
| Example | Description |
|---------|-------------|
| [pricing-decision-ts](./examples/pricing-decision-ts) | Basic pricing workflow (TypeScript) |

## Architecture

```
Client → API (Gin) → Kafka → Consumer → Postgres
```

See [docs/architecture.md](./docs/architecture.md) for details.

## Documentation

- [SDK Reference](./docs/sdk.md) - Python, TypeScript, and curl examples
- [Architecture](./docs/architecture.md) - System design
- [Getting Started](./docs/getting-started.md) - Setup guide
- [API Reference](./docs/api.md) - REST endpoints
- [Concepts](./docs/concepts.md) - Core concepts and terminology

## Tech Stack

- **API**: Go, Gin, SQLC
- **Database**: PostgreSQL
- **Queue**: Kafka (Aiven)
- **SDKs**: Python, TypeScript
- **Docs**: Swagger/OpenAPI

## License

[Elastic License 2.0](./LICENSE)
