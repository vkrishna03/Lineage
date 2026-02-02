# Lineage Core Concepts

Lineage is an epistemic transparency and event lineage system that tracks decisions, assertions, and actions made by both humans and AI systems. This document explains the core concepts.

## Epistemic Transparency

Epistemic transparency means making visible:
- **What** decision/assertion was made
- **Who** made it (human, AI, or automated system)
- **When** it was made
- **Why** it was made (the reasoning or evidence)
- **How confident** the actor was
- **What data** informed the decision

This creates accountability and auditability for both human and AI decision-making.

## Core Entities

### Scopes

A scope defines the boundary for an event chain. Events within a scope form a causally-ordered sequence with hash-chaining for integrity.

Examples:
- A loan application review
- A medical diagnosis session
- A code review process
- A trading decision workflow

### Actors

Actors are entities that generate events. The system supports five actor types:

| Type | Description | Example |
|------|-------------|---------|
| `human` | A person making decisions | Loan officer, doctor, developer |
| `llm` | Large language model | GPT-4, Claude, Gemini |
| `agent` | Autonomous AI agent | AutoGPT, custom agent |
| `service` | Automated service/API | Risk scoring service |
| `tool` | Utility tool | Calculator, code linter |

### Event Types

Event types define the schema and allowed intents for events. Each event type has:
- **Name**: Human-readable identifier
- **Payload Schema**: JSON Schema defining the event payload structure
- **Allowed Intents**: Which intents are valid for this event type

## Events and Intents

Every event has an **intent** that describes its epistemic status:

| Intent | Description | Example |
|--------|-------------|---------|
| `exploration` | Gathering information, no commitment | "Looking at patient history" |
| `suggestion` | Proposing an option without commitment | "Consider prescribing X" |
| `assertion` | Stating a belief or conclusion | "The test result indicates Y" |
| `decision` | Making a choice that affects state | "Approve the loan" |
| `execution` | Carrying out an action | "Loan disbursed" |

Events form a directed acyclic graph (DAG) through parent-child relationships, enabling lineage tracking.

## Artifacts

Artifacts represent data consumed or produced by events. They are content-addressed for deduplication.

Each artifact has:
- **Content Hash**: SHA-256 hash of the content
- **Content Type**: MIME type
- **URI**: Optional storage location
- **Metadata**: Additional context

Artifacts are linked to events with a **role**:
- `input`: Data consumed by the event
- `output`: Data produced by the event

## Scores

Scores quantify aspects of events. The system supports four score types:

| Type | Description |
|------|-------------|
| `confidence` | How certain the actor is (0.0-1.0) |
| `relevance` | How relevant the event is to the scope |
| `reliability` | How reliable the underlying data/source is |
| `agreement` | Level of agreement with prior assertions |

Scores are automatically categorized:

| Value Range | Category |
|-------------|----------|
| 0.00 - 0.19 | `very_low` |
| 0.20 - 0.39 | `low` |
| 0.40 - 0.59 | `moderate` |
| 0.60 - 0.79 | `high` |
| 0.80 - 1.00 | `very_high` |

## Corrections

Events can be corrected through three mechanisms:

| Type | Description | Use Case |
|------|-------------|----------|
| `supersede` | Replaces the corrected event entirely | "Revised assessment replaces original" |
| `amend` | Adds to or modifies the corrected event | "Adding clarification to diagnosis" |
| `retract` | Withdraws the corrected event | "Retracting erroneous recommendation" |

Corrections maintain full history - the original event is never deleted.

## Hash Chaining

Events within a scope are hash-chained:
- Each event contains the hash of the previous event
- Event hashes are computed using RFC 8785 (canonical JSON)
- This creates a tamper-evident audit trail

## Example Workflow

```
1. Create scope: "Loan Application #12345"

2. Human loan officer creates event:
   - Intent: exploration
   - Payload: { "action": "reviewing_application" }

3. Risk scoring service creates event:
   - Intent: assertion
   - Payload: { "risk_score": 720, "factors": [...] }
   - Confidence: 0.92
   - Input artifact: credit_report.json

4. LLM creates event:
   - Intent: suggestion
   - Payload: { "recommendation": "approve", "conditions": [...] }
   - Confidence: 0.78
   - Parent: risk scoring event

5. Human loan officer creates event:
   - Intent: decision
   - Payload: { "decision": "approved", "amount": 50000 }
   - Parent: LLM suggestion event

6. System creates event:
   - Intent: execution
   - Payload: { "disbursement_id": "..." }
   - Output artifact: loan_agreement.pdf
```

## API Quick Reference

### Registry (Setup)
- `POST /api/v1/scopes` - Create scope
- `POST /api/v1/actors` - Register actor
- `POST /api/v1/event-types` - Define event type

### Events
- `POST /api/v1/events` - Create event (via Kafka)
- `GET /api/v1/events/:id` - Get event
- `GET /api/v1/events?scope_id=` - List events in scope
- `GET /api/v1/events/:id/lineage` - Get event lineage

### Scores
- `POST /api/v1/events/:id/scores` - Add score
- `GET /api/v1/events/:id/scores` - Get scores

### Artifacts
- `POST /api/v1/artifacts` - Create artifact
- `GET /api/v1/artifacts/:id` - Get artifact
- `GET /api/v1/artifacts?scope_id=` - List artifacts in scope
- `POST /api/v1/events/:id/artifacts` - Link artifact to event
- `GET /api/v1/events/:id/artifacts` - Get artifacts for event
