# Product Requirements Document (PRD)

## Product Name

**Lineage** — Epistemic Transparency & Event Lineage System

## One-Line Summary

An event-driven system that records how actions and data are produced and changed across humans, AI, and software, so people and systems can later understand where outputs came from and how reliable they might be.

---

## Why This Exists

As AI systems generate, transform, and propagate content at scale, context is rapidly lost. Outputs become detached from their origins, assumptions, confidence levels, and corrections. Without preserved history, humans and AI systems are forced to reason about truth and reliability using incomplete information.

This system exists to **preserve epistemic context** — not to decide truth, but to make truth-seeking possible.

### The Problem

```
Input → [Black Box] → Output
         ↑
    Who made this decision?
    What data informed it?
    How confident were they?
    Was it corrected later?
    Human or AI?
```

### The Solution

```
Input → [Transparent Chain] → Output
              ↓
         Full lineage:
         • Actor (human/AI/service)
         • Intent (suggestion vs decision)
         • Confidence (0.92)
         • Artifacts (input/output data)
         • Corrections (if any)
         • Hash-chained history
```

---

## Core Goal: Epistemic Transparency

Provide a durable, append-only transparency layer that preserves the full causal, generative, and corrective history of actions and data across:

- Humans
- AI models (LLMs)
- Autonomous agents
- Tools and utilities
- APIs and services
- Automated systems

---

## What This System Is Not

| Not This | Why |
|----------|-----|
| Truth oracle | We record claims, not verify them |
| AI detector | Actor type is declared, not inferred |
| Real/fake labeler | That's judgment, not transparency |
| Compliance tool | No enforcement, just visibility |
| Surveillance system | Bounded observation by design |

---

## Design Principles

| Principle | Description | Implementation |
|-----------|-------------|----------------|
| **Transparency Before Judgment** | Record what happened, not what is true | Events capture intent, not correctness |
| **Append-Only History** | Irreversible memory for epistemic trust | Hash-chained events, no deletes |
| **Explicit Uncertainty** | Intent and confidence must be visible | `intent` field + `confidence` scores |
| **Correction Over Erasure** | Truth evolves via correction | `supersede`, `amend`, `retract` |
| **Bounded Observation** | Not everything should be logged | Scopes define boundaries |

---

## Epistemic Boundaries (What Not to Log)

- Private human thought prior to commitment
- Ephemeral drafts and exploration
- Sensitive raw prompts
- Intermediate AI chain-of-thought
- Actions that do not cross belief, decision, or execution boundaries

**Rule of thumb**: If it doesn't cross a commitment boundary, don't log it.

---

## Core Concepts

### Actors

Entities that generate events.

| Type | Description | Example |
|------|-------------|---------|
| `human` | A person making decisions | Loan officer, doctor, developer |
| `llm` | Large language model | GPT-4, Claude, Gemini |
| `agent` | Autonomous AI agent | AutoGPT, custom agents |
| `service` | Automated service/API | Risk scoring service, pricing engine |
| `tool` | Utility tool | Calculator, code linter, validator |

### Events

Immutable records of actions with epistemic context.

| Field | Description |
|-------|-------------|
| `scope_id` | Boundary for the event chain |
| `actor_id` | Who performed the action |
| `event_type_id` | Schema defining payload structure |
| `intent` | Epistemic status of the action |
| `payload` | Event-specific data (JSON) |
| `confidence` | Optional certainty score (0.0-1.0) |
| `event_hash` | SHA-256 of canonical JSON (RFC 8785) |
| `prev_event_hash` | Link to previous event in scope |

### Artifacts

Content-addressed data consumed or produced by events.

| Field | Description |
|-------|-------------|
| `content_hash` | SHA-256 for deduplication |
| `content_type` | MIME type |
| `uri` | Optional storage location |
| `role` | `input` or `output` |

### Scores

Numeric assessments attached to events.

| Type | Description |
|------|-------------|
| `confidence` | How certain the actor is |
| `relevance` | How relevant to the scope |
| `reliability` | How reliable the source/data is |
| `agreement` | Agreement with prior assertions |

**Auto-categorization**:
```
0.00-0.19 → very_low
0.20-0.39 → low
0.40-0.59 → moderate
0.60-0.79 → high
0.80-1.00 → very_high
```

### Lineage

Causal relationships between events forming a DAG (Directed Acyclic Graph).

```
[Risk Assessment] ──derives──▶ [LLM Suggestion] ──derives──▶ [Human Decision]
     (service)                      (llm)                      (human)
   confidence: 0.92              confidence: 0.78            confidence: 0.95
```

---

## Event Intent

The epistemic status of an action.

| Intent | Description | Example |
|--------|-------------|---------|
| `exploration` | Gathering information, no commitment | "Looking at patient history" |
| `suggestion` | Proposing without commitment | "Consider prescribing X" |
| `assertion` | Stating a belief or conclusion | "The test indicates Y" |
| `decision` | Making a choice that affects state | "Approve the loan" |
| `execution` | Carrying out an action | "Loan disbursed" |

**Epistemic weight increases**: exploration → suggestion → assertion → decision → execution

---

## Correction & Retraction

Events are never deleted. Corrections reference and supersede prior events.

| Type | Description | Use Case |
|------|-------------|----------|
| `supersede` | Replaces entirely | "Revised assessment replaces original" |
| `amend` | Adds or modifies | "Adding clarification to diagnosis" |
| `retract` | Withdraws | "Retracting erroneous recommendation" |

```
[Original Event] ◄──corrects── [Correction Event]
                               correction_type: supersede
```

---

## Time Semantics

Events distinguish multiple timestamps for time-relative reasoning.

| Field | Description | Example |
|-------|-------------|---------|
| `observed_at` | When the fact was observed | Data collection time |
| `decided_at` | When the decision was made | Human approval time |
| `ingested_at` | When system recorded it | Write timestamp |

---

## Scoped Trust

Events are scoped to prevent inappropriate generalization.

```
Scope: {
  project: "loan-system",
  domain: "underwriting",
  environment: "production"
}
```

Events in one scope cannot automatically affect trust in another.

---

## System Architecture

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Clients   │────▶│   REST API  │────▶│    Kafka    │
│ (SDK/HTTP)  │     │   (Gin)     │     │  (Aiven)    │
└─────────────┘     └─────────────┘     └──────┬──────┘
                           │                    │
                           │ sync               │ async
                           ▼                    ▼
                    ┌─────────────┐     ┌─────────────┐
                    │  Registry   │     │  Consumer   │
                    │  (CRUD)     │     │  (Writer)   │
                    └──────┬──────┘     └──────┬──────┘
                           │                    │
                           └────────┬───────────┘
                                    ▼
                           ┌─────────────────┐
                           │    Postgres     │
                           │ (Append-only)   │
                           │                 │
                           │ • Events        │
                           │ • Artifacts     │
                           │ • Scores        │
                           │ • Lineage       │
                           └─────────────────┘
```

### Components

| Component | Technology | Purpose |
|-----------|------------|---------|
| API Server | Go + Gin | REST API, validation, Kafka producer |
| Consumer | Go | Kafka consumer, hash computation, persistence |
| Database | PostgreSQL | Append-only event store, lineage graph |
| Message Queue | Kafka (Aiven) | Async event processing, SASL/SSL |

### Hash Chaining (RFC 8785)

```go
hashable := {
    scope_id, actor_id, event_type_id,
    intent, scope_sequence, payload,
    correction_type, corrects_event_id,
    observed_at, decided_at, reason
}
canonical := jsoncanonicalizer.Transform(hashable)
event_hash := sha256(canonical)
```

---

## Use Cases

### 1. AI Agent Decision-Making

Track autonomous agent actions with full accountability.

```
Agent explores options → suggests action → human decides → system executes
     (exploration)         (suggestion)      (decision)     (execution)
```

### 2. Human + AI Collaborative Writing

Distinguish human edits from AI suggestions.

```
AI drafts → Human edits → AI refines → Human approves
  (llm)       (human)        (llm)        (human)
```

### 3. Automated Business Decisions

Audit automated pricing, risk scoring, approvals.

```
Data ingested → Model scores → Rules apply → Action taken
   (service)       (llm)        (service)     (service)
```

### 4. AI Hallucination Debugging

Trace where bad outputs originated.

```
Query lineage: "Where did this claim come from?"
→ Find source event
→ Check confidence score
→ Identify if AI or human
→ See if corrections exist
```

---

## End-to-End Example: Loan Approval

```
Step 1: Credit Score Service
────────────────────────────
Actor: service (Credit Bureau API)
Intent: assertion
Confidence: 0.92
Input Artifact: credit_report.json
Payload: { risk_score: 720, factors: [...] }

        │
        ▼

Step 2: LLM Recommendation
────────────────────────────
Actor: llm (Risk Analyzer GPT)
Intent: suggestion
Confidence: 0.78
Parent: Step 1
Payload: { recommendation: "approve", amount: 50000 }

        │
        ▼

Step 3: Human Decision
────────────────────────────
Actor: human (John Smith, Loan Officer)
Intent: decision
Confidence: 0.95
Parent: Step 2
Payload: { decision: "approved", amount: 45000 }
Reason: "Reduced amount due to market conditions"
Output Artifact: loan_agreement.pdf

        │
        ▼

Step 4: System Execution
────────────────────────────
Actor: service (Disbursement System)
Intent: execution
Parent: Step 3
Payload: { disbursement_id: "..." }
```

**What this enables**:
- Clear separation: AI suggested $50k, human approved $45k
- Accountability: Human made final decision with stated reason
- Auditability: Full chain from data to disbursement
- Debugging: If loan defaults, trace decision lineage

---

## What This Enables

| Capability | Description |
|------------|-------------|
| **Explainability** | Trace any output to its origins |
| **Accountability** | Know who (human/AI) made each decision |
| **Correction visibility** | See how understanding evolved |
| **Uncertainty awareness** | Confidence scores surface doubt |
| **Loop detection** | Identify AI self-reinforcement |

---

## Out of Scope

| Not Building | Reason |
|--------------|--------|
| Truth scoring/ranking | Judgment is downstream, not core |
| Global identity/reputation | Privacy and scope concerns |
| Enforcement/moderation | We record, not enforce |
| Blockchain/tokens | Unnecessary complexity |

---

## Success Metrics

| Metric | Target |
|--------|--------|
| Outputs are explainable | Any output traceable to source |
| Corrections are visible | <1 min to find correction history |
| AI agents reason about uncertainty | Confidence accessible via API |
| Trust discussions rely on evidence | Lineage queryable |

---

## Implementation Status

| Phase | Description | Status |
|-------|-------------|--------|
| Phase 1 | Event protocol & schema | ✅ Complete |
| Phase 2 | Append-only storage & validation | ✅ Complete |
| Phase 3 | Lineage querying | ✅ Complete |
| Phase 4 | Artifacts & Scores | ✅ Complete |
| Phase 5 | SDKs & adoption | 🔄 Planned |

### What's Built

- **API Server**: REST endpoints for registry and events
- **Kafka Integration**: Async event processing with SASL/SSL
- **Hash Chaining**: RFC 8785 canonical JSON, SHA-256
- **Artifacts**: Content-addressed with deduplication
- **Scores**: Confidence, relevance, reliability, agreement
- **Lineage**: Parent/child relationships, DAG traversal
- **Validation**: JSON Schema for payloads, intent validation
- **Documentation**: OpenAPI/Swagger, concepts guide

### API Endpoints

See [API Reference](./api.md) for full documentation.

| Category | Endpoints |
|----------|-----------|
| Registry | `/scopes`, `/actors`, `/event-types` |
| Events | `/events`, `/events/:id/lineage` |
| Scores | `/events/:id/scores` |
| Artifacts | `/artifacts`, `/events/:id/artifacts` |

---

## Next Steps

1. **SDKs** — Python, TypeScript, Go clients
2. **Frontend** — Lineage visualization UI
3. **Authentication** — API keys, JWT
4. **Advanced Queries** — Recursive lineage traversal
5. **Analytics** — Aggregations by actor, intent, time
