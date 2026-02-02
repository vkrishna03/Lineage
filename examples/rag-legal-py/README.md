# Legal RAG Pipeline with Lineage

A RAG (Retrieval-Augmented Generation) pipeline for legal document analysis that tracks full provenance from source documents to final legal opinion.

## What It Demonstrates

- **Document ingestion** with content hashing and artifact tracking
- **RAG retrieval** with provenance linking to source documents
- **AI analysis** (suggestion intent) with confidence scores
- **Human lawyer review** (decision intent) validating AI output
- **Execution tracking** for audit trail
- **Full lineage chain** from documents → retrieval → analysis → decision → action

## The Scenario

A compliance team needs to determine if their employment contract's non-compete clause is enforceable in California. The pipeline:

1. Ingests legal documents (contracts, statutes, case law)
2. Retrieves relevant documents based on the query
3. AI generates legal analysis with risk assessment
4. Human lawyer reviews and approves
5. System executes the recommendation

## Running

```bash
# Install dependencies
uv sync

# Run with mock responses (no API key needed)
uv run python main.py

# Run with real OpenAI (requires OPENAI_API_KEY)
OPENAI_API_KEY=your-key uv run python main.py
```

**Note**: Requires Lineage API server running at `http://localhost:8080`

## Event Chain

```
[document_ingestion] x3     (assertion, 0.99) - Facts about documents
         ↓
[document_retrieval]        (exploration, 0.85) - Gathering information
         ↓
[legal_analysis]            (suggestion, 0.72) - AI proposes interpretation
         ↓
[legal_review]              (decision, 0.95) - Human validates
         ↓
[recommendation_execution]  (execution, 1.0) - Action taken
```

## Actors

| Actor | Type | Role |
|-------|------|------|
| Document Ingestion Service | service | Ingests and indexes documents |
| Vector Search | tool | Retrieves relevant documents |
| GPT-4 Legal Analyst | llm | Generates legal analysis |
| Sarah Chen, Senior Counsel | human | Reviews and approves |
| Compliance System | service | Executes recommendations |

## Key Patterns

### Artifact Tracking
Documents are ingested with content hashes, allowing verification that analysis is based on specific document versions.

### Confidence Progression
- **Ingestion**: 0.99 (facts are certain)
- **Retrieval**: 0.85 (relevance is probabilistic)
- **AI Analysis**: 0.72 (interpretation is uncertain)
- **Human Decision**: 0.95 (human judgment is high confidence)
- **Execution**: 1.0 (action is certain)

### Lineage Links
Each event links to its parent, creating a traceable chain:
- Retrieval links to ingested documents
- Analysis links to retrieval
- Decision links to analysis
- Execution links to decision

## Sample Output

```
============================================================
Legal RAG Pipeline with Lineage Tracking
============================================================

[Phase 1] Document Ingestion
----------------------------------------
  Ingested: Employment Contract Template (hash: sha256:a1b2c3d4...)
  Ingested: California Non-Compete Law Summary (hash: sha256:e5f6g7h8...)
  Ingested: Recent Court Ruling: TechCorp v. Smith (2024) (hash: sha256:i9j0k1l2...)

[Phase 2] Processing Query
----------------------------------------
Case ID: CASE-2024-001
Query: Is the non-compete clause in our California employment contracts enforceable?

[Phase 3] Document Retrieval
----------------------------------------
  Retrieved 3 relevant documents

[Phase 4] AI Analysis
----------------------------------------
  Generated analysis (risk: high)

[Phase 5] Lawyer Review
----------------------------------------
  Lawyer review: APPROVED

[Phase 6] Execute Recommendation
----------------------------------------
  Executed: do_not_enforce

============================================================
Pipeline Complete - Lineage Tracked
============================================================
```

## Why This Matters

In legal and compliance contexts, you need to:

1. **Prove provenance**: Which documents informed the decision?
2. **Track AI involvement**: What did the AI suggest vs. what humans decided?
3. **Audit trail**: Who approved what, when, and why?
4. **Version control**: Was the analysis based on current documents?

Lineage provides all of this automatically through the event chain.
