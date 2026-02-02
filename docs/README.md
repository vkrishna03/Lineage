# Documentation

Detailed documentation for the Lineage project.

## Contents

| Document | Description |
|----------|-------------|
| [SDK Reference](./sdk.md) | Python, TypeScript SDKs and curl examples |
| [Concepts](./concepts.md) | Core concepts and terminology |
| [Architecture](./architecture.md) | System design, components, data model |
| [Getting Started](./getting-started.md) | Local setup and development |
| [API Reference](./api.md) | REST endpoints and OpenAPI spec |
| [PRD](./prd.md) | Product requirements and vision |

## SDKs

- [Python SDK](../sdks/python) - `pip install lineage-sdk`
- [TypeScript SDK](../sdks/typescript) - `npm install lineage-sdk`

## Examples

### Python
- [pricing-decision-py](../examples/pricing-decision-py) - Basic pricing workflow
- [langgraph-support-py](../examples/langgraph-support-py) - LangGraph customer support agent
- [rag-legal-py](../examples/rag-legal-py) - RAG pipeline for legal documents
- [fastapi-service-py](../examples/fastapi-service-py) - FastAPI service with approval workflow

### TypeScript
- [pricing-decision-ts](../examples/pricing-decision-ts) - Basic pricing workflow

## Additional Resources

- [Database Schema](../schema.dbml) - DBML schema definition
- [Swagger UI](http://localhost:8080/swagger/index.html) - Interactive API docs (when running)
