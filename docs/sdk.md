# SDK Reference

Lineage provides SDKs for Python and TypeScript, plus direct API access via curl.

## Quick Comparison

| Feature | Python | TypeScript | curl |
|---------|--------|------------|------|
| Simple API | `lineage.init()`, `emit()`, `@track` | `lineage.init()`, `emit()`, `track()` | N/A |
| Low-level Client | `LineageClient` | `LineageClient` | Direct HTTP |
| Type Safety | Pydantic models | TypeScript types | None |
| Async Support | Sync (with wait) | Async/await | N/A |

## Python SDK

### Installation

```bash
pip install lineage-sdk
# or with uv
uv add lineage-sdk
```

### Simple API

```python
import lineage

# Initialize
lineage.init(
    project="my-app",
    actor_name="my-service",
    actor_type="service"
)

# Emit events
lineage.emit("data_ingestion", "assertion", {"data": "loaded"}, confidence=0.99)

# Track functions
@lineage.track("recommendation", intent="suggestion", actor=("llm", "GPT-4"))
def recommend(data):
    return {"price": 26.99, "confidence": 0.85}

# Context manager
with lineage.span("processing", "assertion") as span:
    span.payload = {"rows": 1000}
    span.confidence = 0.99
```

### Low-level Client

```python
from lineage import LineageClient, ActorType, Intent

client = LineageClient(base_url="http://localhost:8080")

scope = client.scopes.create(project="my-project")
actor = client.actors.create(type=ActorType.LLM, name="My LLM")
```

[Full Python SDK documentation](../sdks/python/README.md)

---

## TypeScript SDK

### Installation

```bash
npm install lineage-sdk
```

### Simple API

```typescript
import * as lineage from 'lineage-sdk';

// Initialize
await lineage.init({
  project: 'my-app',
  actorName: 'my-service',
  actorType: 'service'
});

// Emit events
await lineage.emit('data_ingestion', 'assertion', { data: 'loaded' }, { confidence: 0.99 });

// Track functions
const recommend = lineage.track('recommendation', 'suggestion', {
  actor: ['llm', 'GPT-4']
})(async (data) => {
  return { price: 26.99, confidence: 0.85 };
});
```

### Low-level Client

```typescript
import { LineageClient } from 'lineage-sdk';

const client = new LineageClient({ baseUrl: 'http://localhost:8080' });

const scope = await client.scopes.create({ project: 'my-project' });
const actor = await client.actors.create({ type: 'llm', name: 'My LLM' });
```

[Full TypeScript SDK documentation](../sdks/typescript/README.md)

---

## curl Examples

Direct API access without SDKs.

### Setup Registry

```bash
# Create a scope
curl -X POST http://localhost:8080/api/v1/scopes \
  -H "Content-Type: application/json" \
  -d '{
    "project": "my-app",
    "domain": "pricing",
    "environment": "production"
  }'
# Response: {"id": "550e8400-...", "project": "my-app", ...}

# Create actors
curl -X POST http://localhost:8080/api/v1/actors \
  -H "Content-Type: application/json" \
  -d '{
    "type": "service",
    "name": "Data Pipeline"
  }'

curl -X POST http://localhost:8080/api/v1/actors \
  -H "Content-Type: application/json" \
  -d '{
    "type": "llm",
    "name": "GPT-4",
    "metadata": {"model": "gpt-4-turbo", "temperature": 0.7}
  }'

curl -X POST http://localhost:8080/api/v1/actors \
  -H "Content-Type: application/json" \
  -d '{
    "type": "human",
    "name": "Alice Smith",
    "metadata": {"role": "Manager"}
  }'

# Create event type
curl -X POST http://localhost:8080/api/v1/event-types \
  -H "Content-Type: application/json" \
  -d '{
    "name": "pricing_workflow",
    "version": "1.0",
    "allowed_intents": ["assertion", "suggestion", "decision", "execution"],
    "payload_schema": {
      "type": "object",
      "properties": {
        "product_id": {"type": "string"},
        "price": {"type": "number"}
      }
    }
  }'
```

### Create Events Chain

```bash
# Store IDs from previous responses
SCOPE_ID="your-scope-id"
SERVICE_ACTOR_ID="your-service-actor-id"
LLM_ACTOR_ID="your-llm-actor-id"
HUMAN_ACTOR_ID="your-human-actor-id"
EVENT_TYPE_ID="your-event-type-id"

# Step 1: Service asserts data (high confidence)
curl -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -d '{
    "scope_id": "'$SCOPE_ID'",
    "actor_id": "'$SERVICE_ACTOR_ID'",
    "event_type_id": "'$EVENT_TYPE_ID'",
    "intent": "assertion",
    "confidence": 0.99,
    "payload": {
      "product_id": "SKU-123",
      "current_price": 29.99,
      "units_sold": 1500
    }
  }'
# Response: {"status": "accepted", "message": "Event queued for processing"}

# Wait for async processing (events go through Kafka)
sleep 2

# Get the event ID
ASSERTION_ID=$(curl -s "http://localhost:8080/api/v1/events?scope_id=$SCOPE_ID" | jq -r '.events[0].id')

# Step 2: LLM suggests price (moderate confidence)
curl -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -d '{
    "scope_id": "'$SCOPE_ID'",
    "actor_id": "'$LLM_ACTOR_ID'",
    "event_type_id": "'$EVENT_TYPE_ID'",
    "intent": "suggestion",
    "confidence": 0.72,
    "parent_event_ids": ["'$ASSERTION_ID'"],
    "payload": {
      "product_id": "SKU-123",
      "recommended_price": 26.99,
      "reasoning": "Competitor analysis suggests price reduction"
    }
  }'

sleep 2
SUGGESTION_ID=$(curl -s "http://localhost:8080/api/v1/events?scope_id=$SCOPE_ID" | jq -r '.events[] | select(.intent=="suggestion") | .id')

# Step 3: Human decides (high confidence)
curl -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -d '{
    "scope_id": "'$SCOPE_ID'",
    "actor_id": "'$HUMAN_ACTOR_ID'",
    "event_type_id": "'$EVENT_TYPE_ID'",
    "intent": "decision",
    "confidence": 0.88,
    "parent_event_ids": ["'$SUGGESTION_ID'"],
    "reason": "Accepting with margin adjustment",
    "payload": {
      "product_id": "SKU-123",
      "approved_price": 27.99
    }
  }'

sleep 2
DECISION_ID=$(curl -s "http://localhost:8080/api/v1/events?scope_id=$SCOPE_ID" | jq -r '.events[] | select(.intent=="decision") | .id')

# Step 4: Service executes (certain)
curl -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -d '{
    "scope_id": "'$SCOPE_ID'",
    "actor_id": "'$SERVICE_ACTOR_ID'",
    "event_type_id": "'$EVENT_TYPE_ID'",
    "intent": "execution",
    "confidence": 1.0,
    "parent_event_ids": ["'$DECISION_ID'"],
    "payload": {
      "product_id": "SKU-123",
      "old_price": 29.99,
      "new_price": 27.99,
      "effective_at": "2024-01-15T00:00:00Z"
    }
  }'
```

### Query Lineage

```bash
# List all events in scope
curl "http://localhost:8080/api/v1/events?scope_id=$SCOPE_ID" | jq

# Get event details
curl "http://localhost:8080/api/v1/events/$DECISION_ID" | jq

# Get lineage (parents and children)
curl "http://localhost:8080/api/v1/events/$DECISION_ID/lineage" | jq
# Response:
# {
#   "event_id": "...",
#   "parents": [{ "id": "...", "intent": "suggestion", ... }],
#   "children": [{ "id": "...", "intent": "execution", ... }]
# }

# Get scores for an event
curl "http://localhost:8080/api/v1/events/$DECISION_ID/scores" | jq
```

### Add Scores

```bash
# Add a reliability score to flag weak assumption
curl -X POST "http://localhost:8080/api/v1/events/$SUGGESTION_ID/scores" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "reliability",
    "value": 0.45,
    "scored_by": "'$HUMAN_ACTOR_ID'",
    "reason": "Competitor reaction assumption is weak"
  }'
```

### Create Artifacts

```bash
# Create an artifact
curl -X POST http://localhost:8080/api/v1/artifacts \
  -H "Content-Type: application/json" \
  -d '{
    "scope_id": "'$SCOPE_ID'",
    "content_hash": "sha256:abc123def456",
    "content_type": "application/json",
    "uri": "s3://bucket/sales-data.json",
    "metadata": {"rows": 50000}
  }'

ARTIFACT_ID="your-artifact-id"

# Link artifact to event
curl -X POST "http://localhost:8080/api/v1/events/$ASSERTION_ID/artifacts" \
  -H "Content-Type: application/json" \
  -d '{
    "artifact_id": "'$ARTIFACT_ID'",
    "role": "input"
  }'

# Get artifacts for event
curl "http://localhost:8080/api/v1/events/$ASSERTION_ID/artifacts" | jq
```

### Corrections

```bash
# Amend an event (partial correction)
curl -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -d '{
    "scope_id": "'$SCOPE_ID'",
    "actor_id": "'$HUMAN_ACTOR_ID'",
    "event_type_id": "'$EVENT_TYPE_ID'",
    "intent": "assertion",
    "correction_type": "amend",
    "corrects_event_id": "'$SUGGESTION_ID'",
    "reason": "Correcting competitor reaction assumption",
    "payload": {
      "amendment": "competitor_reaction",
      "original": "30 days",
      "corrected": "14 days"
    }
  }'

# Supersede an event (full replacement)
curl -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -d '{
    "scope_id": "'$SCOPE_ID'",
    "actor_id": "'$LLM_ACTOR_ID'",
    "event_type_id": "'$EVENT_TYPE_ID'",
    "intent": "suggestion",
    "correction_type": "supersede",
    "corrects_event_id": "'$SUGGESTION_ID'",
    "reason": "New analysis with updated data",
    "payload": {
      "product_id": "SKU-123",
      "recommended_price": 28.49,
      "reasoning": "Updated competitor data"
    }
  }'

# Retract an event (invalidate)
curl -X POST http://localhost:8080/api/v1/events \
  -H "Content-Type: application/json" \
  -d '{
    "scope_id": "'$SCOPE_ID'",
    "actor_id": "'$HUMAN_ACTOR_ID'",
    "event_type_id": "'$EVENT_TYPE_ID'",
    "intent": "assertion",
    "correction_type": "retract",
    "corrects_event_id": "'$SUGGESTION_ID'",
    "reason": "Analysis was based on incorrect data",
    "payload": {"retracted": true}
  }'
```

### Health Check

```bash
curl http://localhost:8080/health | jq
# Response: {"status": "ok", "services": {"kafka": "ok", "postgres": "ok"}}
```

---

## Intent Reference

| Intent | Description | Confidence Range | Example |
|--------|-------------|------------------|---------|
| `exploration` | Gathering information | 0.7-0.9 | Analyzing data |
| `suggestion` | Proposing without commitment | 0.5-0.8 | AI recommendation |
| `assertion` | Stating a fact | 0.8-1.0 | Data ingestion |
| `decision` | Making a choice | 0.8-0.95 | Human approval |
| `execution` | Taking action | 1.0 | Applying change |

## Actor Types

| Type | Use For |
|------|---------|
| `human` | Human users, reviewers, managers |
| `llm` | Language models (GPT-4, Claude, Gemini) |
| `agent` | Autonomous agents (LangGraph, CrewAI) |
| `service` | Backend services, APIs, microservices |
| `tool` | Utility tools, calculators, validators |

## Score Types

| Type | Description |
|------|-------------|
| `confidence` | How certain the actor is (auto-added) |
| `relevance` | How relevant to the context |
| `reliability` | How reliable the source/data is |
| `agreement` | Agreement with prior assertions |
