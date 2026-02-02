# AI-Assisted Pricing Decision Example

This example demonstrates the complete workflow from the Lineage PRD:

1. **Service ingests sales data** - Data pipeline asserts sales metrics
2. **AI analyzes and recommends** - LLM suggests price reduction with assumptions
3. **Human reviews and decides** - Manager accepts with modifications
4. **Human supersedes weak assumption** - Manager amends AI's weak assumption
5. **Service executes change** - Pricing engine executes approved change

## Prerequisites

1. Start the Lineage API server:
   ```bash
   cd apps/api
   make run-api &
   make run-consumer &
   ```

2. Ensure Kafka and PostgreSQL are running (see main README)

## Running

```bash
# Using uv
uv sync
uv run python main.py

# Or directly
uv run main.py
```

## What This Demonstrates

### Epistemic Transparency
- **Intent tracking**: Each event has a clear intent (assertion, suggestion, decision, execution)
- **Confidence scores**: AI suggestion has 0.72 confidence; human decision has 0.88
- **Reason capture**: Human provides reason for adjusting AI recommendation

### Event Lineage
- Events are linked via `parent_event_ids`
- Full decision chain is queryable
- Corrections reference the event being corrected

### Corrections
- Human adds reliability score to flag weak AI assumption
- Human creates AMEND correction to fix the assumption
- Original events remain immutable; corrections are new events

### Actor Types
- `SERVICE`: Data pipeline, pricing engine
- `LLM`: Pricing optimizer AI
- `HUMAN`: Pricing manager

## Expected Output

```
============================================================
AI-Assisted Pricing Decision Workflow
============================================================

API Status: healthy

--- Setting up registry ---
Scope: ecommerce/pricing
Actors: Sales Data Pipeline, Pricing Optimizer GPT, Sarah Chen, Pricing Engine

============================================================
STEP 1: Service ingests sales data
============================================================
✓ Data ingestion event created
  Actor: Sales Data Pipeline (service)
  Intent: assertion
  Confidence: 0.99 (verified data)

============================================================
STEP 2: AI model analyzes and recommends pricing
============================================================
✓ AI recommendation created
  Actor: Pricing Optimizer GPT (llm)
  Intent: suggestion
  Confidence: 0.72 (moderate - assumptions involved)
  Recommended: $26.99 (from $29.99)

============================================================
STEP 3: Human reviews and makes decision
============================================================
✓ Human decision created
  Actor: Sarah Chen (human)
  Intent: decision
  Confidence: 0.88
  Approved: $27.99 (adjusted from AI's $26.99)

============================================================
STEP 4: Human supersedes weak assumption
============================================================
✓ Correction event created
  Type: amend
  Corrects: AI suggestion event

============================================================
STEP 5: Service executes approved pricing change
============================================================
✓ Execution event created
  Price changed: $29.99 → $27.99

============================================================
LINEAGE: Full decision trail
============================================================

Decision chain (newest to oldest):
→ [service] Pricing Engine: execution (confidence: 1.0)
  → [human] Sarah Chen: decision (confidence: 0.88)
    → [llm] Pricing Optimizer GPT: suggestion (confidence: 0.72)
      → [service] Sales Data Pipeline: assertion (confidence: 0.99)

============================================================
✓ Workflow complete - full epistemic transparency preserved
============================================================
```

## Key Takeaways

1. **Every decision is traceable** - From execution back to original data
2. **AI suggestions are clearly marked** - Intent: suggestion, not decision
3. **Human oversight is recorded** - Manager's adjustments are captured
4. **Weak assumptions are correctable** - Without losing original context
5. **Confidence propagates** - Lower confidence at suggestion, higher at verified data
