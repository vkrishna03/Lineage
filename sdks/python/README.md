# Lineage Python SDK

Epistemic transparency for AI systems - track decisions, suggestions, and actions with full lineage.

## Installation

```bash
uv add lineage-sdk
# or
pip install lineage-sdk
```

## Quick Start

```python
import lineage

# Initialize once
lineage.init(project="my-app", actor_name="my-service", actor_type="service")

# Emit events with one line
lineage.emit("data_ingestion", "assertion", {"data": "loaded"}, confidence=0.99)

# Or use decorators
@lineage.track("recommendation", intent="suggestion")
def recommend(data):
    return {"price": 26.99, "confidence": 0.85}
```

## Simple API

### Initialize

```python
import lineage

lineage.init(
    project="my-app",                  # Required: project name
    base_url="http://localhost:8080",  # Lineage API server
    domain="pricing",                  # Optional: domain
    environment="production",          # Optional: environment
    actor_name="my-service",           # Default actor name
    actor_type="service",              # human, llm, agent, service, tool
)
```

### Emit Events

```python
# Simple emit
lineage.emit(
    "event_type_name",      # Auto-created if needed
    "assertion",            # intent: exploration, suggestion, assertion, decision, execution
    {"key": "value"},       # payload
    confidence=0.95,        # optional confidence score
    actor=("llm", "GPT-4"), # optional actor override (type, name)
    parent=previous_event,  # optional parent for lineage
)
```

### Track Functions with Decorator

```python
@lineage.track("recommendation", intent="suggestion", actor=("llm", "Pricing AI"))
def recommend_price(data):
    # Return value becomes payload
    # 'confidence' key is auto-extracted
    return {
        "recommended_price": 26.99,
        "reasoning": "Market analysis",
        "confidence": 0.72
    }

result = recommend_price(input_data)  # Event auto-emitted
```

### Track Code Blocks with Span

```python
with lineage.span("data_processing", "assertion", actor=("service", "ETL")) as span:
    result = process_data()
    span.payload = {"rows_processed": 1000}
    span.confidence = 0.99

# Access the created event
event = span.event
```

### Intents

| Intent | Use When |
|--------|----------|
| `exploration` | Gathering information, research |
| `suggestion` | AI/LLM proposing an option |
| `assertion` | Stating a fact, data ingestion |
| `decision` | Human/system making a choice |
| `execution` | Taking action, applying changes |

### Actor Types

| Type | Use For |
|------|---------|
| `human` | Human users, reviewers |
| `llm` | Language models (GPT-4, Claude, etc.) |
| `agent` | Autonomous agents (LangGraph, CrewAI) |
| `service` | Backend services, APIs |
| `tool` | Tools, functions called by agents |

## Full Example

```python
"""AI-Assisted Pricing Decision"""
import lineage

lineage.init(project="ecommerce", domain="pricing")

# Service ingests data
lineage.emit(
    "data_ingestion", "assertion",
    {"product": "SKU-123", "price": 29.99},
    confidence=0.99,
    actor=("service", "Data Pipeline")
)

# AI recommends
@lineage.track("recommendation", intent="suggestion", actor=("llm", "Pricing AI"))
def recommend_price(data):
    return {"recommended_price": 26.99, "confidence": 0.72}

rec = recommend_price({"product": "SKU-123"})

# Human decides
with lineage.span("recommendation", "decision", actor=("human", "Manager")) as span:
    span.payload = {"approved_price": 27.99}
    span.confidence = 0.88

# Service executes
lineage.emit(
    "execution", "execution",
    {"new_price": 27.99},
    confidence=1.0,
    actor=("service", "Pricing Engine"),
    parent=span.event
)
```

## Low-Level API

For advanced use cases, access the underlying client:

```python
from lineage import LineageClient, ActorType, Intent

client = LineageClient(base_url="http://localhost:8080")

# Full control over resources
scope = client.scopes.create(project="my-project")
actor = client.actors.create(type=ActorType.LLM, name="My LLM")
event_type = client.event_types.create(
    name="custom_event",
    version="1.0",
    allowed_intents=[Intent.SUGGESTION]
)

client.events.create(
    scope_id=scope.id,
    actor_id=actor.id,
    event_type_id=event_type.id,
    intent=Intent.SUGGESTION,
    payload={"data": "value"}
)
```

## Framework Integrations (Coming Soon)

- **FastAPI**: Middleware for automatic request tracking
- **LangChain**: Callback handler for LLM call tracking
- **LangGraph**: Node wrapper for graph execution tracking

## Development

```bash
uv sync --extra dev
uv run pytest
uv run ruff check .
```

## License

Elastic-2.0
