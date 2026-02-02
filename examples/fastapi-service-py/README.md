# FastAPI Service with Lineage

A production-ready FastAPI service demonstrating Lineage integration for AI-assisted pricing with human approval workflow.

## What It Demonstrates

- **Startup initialization** with FastAPI lifespan
- **Request-scoped tracking** via middleware
- **Multiple actors** in a single request flow (service → AI → human → service)
- **Approval workflow** with pending state management
- **Human override** of AI recommendations
- **Full audit trail** for compliance

## The Scenario

A pricing API that:
1. Receives product data and competitor prices
2. AI model recommends a price (suggestion)
3. Human manager reviews and approves/rejects (decision)
4. System applies the price change (execution)

## Running

```bash
# Install dependencies
uv sync

# Start the FastAPI server
uv run uvicorn main:app --reload

# API docs at http://localhost:8000/docs
```

**Note**: Requires Lineage API server running at `http://localhost:8080`

## API Endpoints

### POST /api/v1/price/recommend
Get AI price recommendation.

```bash
curl -X POST http://localhost:8000/api/v1/price/recommend \
  -H "Content-Type: application/json" \
  -d '{
    "product_id": "SKU-123",
    "current_price": 29.99,
    "competitor_prices": [27.99, 28.50, 31.00],
    "inventory_level": 100
  }'
```

Response:
```json
{
  "product_id": "SKU-123",
  "recommended_price": 28.49,
  "confidence": 0.75,
  "reasoning": "Price is 105% of competitor average. Recommend 5% reduction.",
  "request_id": "a1b2c3d4"
}
```

### POST /api/v1/price/approve
Approve or reject recommendation.

```bash
# Approve as-is
curl -X POST http://localhost:8000/api/v1/price/approve \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "a1b2c3d4",
    "approved": true,
    "reviewer_notes": "Approved - competitive pricing needed"
  }'

# Approve with override
curl -X POST http://localhost:8000/api/v1/price/approve \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "a1b2c3d4",
    "approved": true,
    "override_price": 27.99,
    "reviewer_notes": "Adjusted to match lowest competitor"
  }'

# Reject
curl -X POST http://localhost:8000/api/v1/price/approve \
  -H "Content-Type: application/json" \
  -d '{
    "request_id": "a1b2c3d4",
    "approved": false,
    "reviewer_notes": "Margins too thin - maintain current price"
  }'
```

### GET /api/v1/price/pending
List pending approvals.

```bash
curl http://localhost:8000/api/v1/price/pending
```

## Event Chain

```
[price_data_ingestion]    (assertion, 0.99)  - Input data recorded
         ↓
[price_recommendation]    (suggestion, 0.7)  - AI recommends
         ↓
[price_decision]          (decision, 0.95)   - Human approves/rejects
         ↓
[price_execution]         (execution, 1.0)   - Price applied
```

## Actors

| Actor | Type | Events |
|-------|------|--------|
| Pricing API | service | Data ingestion, execution |
| Pricing AI Model | llm | Recommendations |
| Pricing Manager | human | Approvals |

## Key Patterns

### Lifespan Initialization
```python
@asynccontextmanager
async def lifespan(app: FastAPI):
    await lineage.init(
        project="pricing-api",
        domain="pricing",
        base_url="http://localhost:8080",
    )
    yield

app = FastAPI(lifespan=lifespan)
```

### Request-Scoped Tracking
```python
@app.middleware("http")
async def add_request_id(request: Request, call_next):
    request.state.request_id = str(uuid4())[:8]
    response = await call_next(request)
    return response
```

### Parent Event Linking
```python
# Data ingestion
data_event = lineage.emit("data_ingestion", "assertion", {...})

# AI suggestion links to data
suggestion_event = lineage.emit(
    "recommendation", "suggestion", {...},
    parent=data_event.id,
)

# Human decision links to suggestion
decision_event = lineage.emit(
    "decision", "decision", {...},
    parent=suggestion_event.id,
)
```

### Confidence Progression
- **Data ingestion**: 0.99 (facts)
- **AI recommendation**: 0.6-0.82 (varies by model confidence)
- **Human decision**: 0.95 (human judgment)
- **Execution**: 1.0 (action taken)

## Full Demo Flow

```bash
# 1. Start the server
uv run uvicorn main:app --reload

# 2. Request a recommendation
curl -X POST http://localhost:8000/api/v1/price/recommend \
  -H "Content-Type: application/json" \
  -d '{"product_id": "SKU-123", "current_price": 29.99, "competitor_prices": [27.99, 28.50]}'

# Note the request_id in the response

# 3. Approve with override
curl -X POST http://localhost:8000/api/v1/price/approve \
  -H "Content-Type: application/json" \
  -d '{"request_id": "YOUR_REQUEST_ID", "approved": true, "override_price": 27.99}'

# 4. Check Lineage events
curl "http://localhost:8080/api/v1/events?scope_id=YOUR_SCOPE_ID" | jq
```

## Production Considerations

1. **State Management**: Replace `pending_approvals` dict with Redis or database
2. **Authentication**: Add JWT/API key validation
3. **Rate Limiting**: Protect AI recommendation endpoint
4. **Async Processing**: Consider background tasks for execution
5. **Error Handling**: Add comprehensive error tracking with Lineage

## Why This Pattern?

In production AI systems, you need:
- **Audit trail**: Who approved what price change, when?
- **AI transparency**: What did the model recommend vs. final outcome?
- **Override tracking**: When do humans override AI?
- **Rollback capability**: Trace any price to its decision chain
