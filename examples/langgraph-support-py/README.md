# LangGraph Customer Support Agent

A production-style customer support agent using LangGraph with full Lineage tracking.

## What This Demonstrates

1. **LangGraph Integration** - State graph with conditional routing
2. **Multi-Actor Workflow** - Agent, LLM, Human, Service
3. **Intent Progression** - exploration → suggestion → decision → execution
4. **Human-in-the-Loop** - Escalation for negative sentiment
5. **Full Lineage** - Every decision tracked and auditable

## Architecture

```
Customer Query
      │
      ▼
┌─────────────────┐
│ Sentiment       │  Intent: exploration
│ Analysis        │  Actor: agent
│ (exploration)   │  Confidence: 0.85
└────────┬────────┘
         │
         ▼
┌─────────────────┐
│ Generate        │  Intent: suggestion
│ Response        │  Actor: llm
│ (suggestion)    │  Confidence: 0.72
└────────┬────────┘
         │
    ┌────┴────┐
    │         │
    ▼         ▼
Positive   Negative
    │         │
    │    ┌────┴────────┐
    │    │ Human       │  Intent: decision
    │    │ Review      │  Actor: human
    │    │ (decision)  │  Confidence: 0.95
    │    └────┬────────┘
    │         │
    └────┬────┘
         │
         ▼
┌─────────────────┐
│ Send Response   │  Intent: execution
│ (execution)     │  Actor: service
│                 │  Confidence: 1.0
└─────────────────┘
```

## Running

### Prerequisites

1. Lineage API server running:
   ```bash
   cd apps/api
   make run-api &
   make run-consumer &
   ```

2. (Optional) OpenAI API key for real LLM:
   ```bash
   export OPENAI_API_KEY=sk-...
   ```
   Without this, the example uses mock responses.

### Run the Example

```bash
cd examples/langgraph-support-py
uv sync
uv run python main.py
```

## Sample Output

```
============================================================
LangGraph Customer Support with Lineage Tracking
============================================================

Lineage project: customer-support-abc123

============================================================
TEST CASE 1: Simple Query (Auto-Response)
============================================================

Processing Ticket: TICKET-001
Query: Hi, I'd like to know what time your store closes on weekends...

Sentiment: positive
Escalated: False

Final Response:
----------------------------------------
Thank you for contacting us. I understand your concern...
----------------------------------------

============================================================
TEST CASE 2: Angry Customer (Human Escalation)
============================================================

Processing Ticket: TICKET-002
Query: I am FURIOUS! My order arrived broken...

Sentiment: negative
Escalated: True

Final Response:
----------------------------------------
I sincerely apologize for the frustration...
I'm also adding a 20% discount to your next order...
----------------------------------------

============================================================
LINEAGE SUMMARY
============================================================

Total events tracked: 8

Event chain:
  [agent] exploration: Support Agent (conf: 0.85)
  [llm] suggestion: GPT-4o-mini (conf: 0.72)
  [service] execution: Email Service (conf: 1.00)
  [agent] exploration: Support Agent (conf: 0.85)
  [llm] suggestion: GPT-4o-mini (conf: 0.72)
  [human] decision: Sarah Chen, Support Lead (conf: 0.95)
  [service] execution: Email Service (conf: 1.00)
```

## Key Patterns

### 1. Intent Selection

| Stage | Intent | Why |
|-------|--------|-----|
| Sentiment Analysis | `exploration` | Gathering info, no commitment |
| Generate Response | `suggestion` | AI proposing, not final |
| Human Review | `decision` | Human authority, final call |
| Send Response | `execution` | Action taken |

### 2. Confidence Scores

- **Exploration**: 0.85 - Analysis is fairly reliable
- **Suggestion**: 0.72 - AI recommendation, lower confidence
- **Decision**: 0.95 - Human reviewed, high confidence
- **Execution**: 1.0 - Deterministic action

### 3. Lineage Linking

Each event links to its parent:
```python
event = lineage.emit(
    "response_draft",
    "suggestion",
    payload,
    parent=state["lineage_events"][-1]  # Link to previous event
)
```

### 4. Actor Types

- `agent` - LangGraph agent (orchestration)
- `llm` - Language model (GPT-4)
- `human` - Human reviewer
- `service` - Backend service (email)

## Production Considerations

1. **Human Review Integration**: Replace mock with webhook/queue
2. **Error Handling**: Track failed events with corrections
3. **Async Processing**: Use Lineage's Kafka-backed async events
4. **Monitoring**: Query lineage for escalation rates, confidence trends
