# Pricing Decision Workflow (TypeScript)

A pricing workflow example demonstrating Lineage tracking with the TypeScript SDK.

## What It Demonstrates

- **Multi-actor workflow**: service → LLM → human → service
- **Intent progression**: assertion → suggestion → decision → execution
- **Confidence scores**: Different confidence levels for each actor type
- **Event linking**: Parent-child relationships for lineage

## Running

```bash
# Install dependencies
npm install

# Run the example
npm start
# or
npx tsx main.ts
```

**Note**: Requires Lineage API server running at `http://localhost:8080`

## Event Chain

```
[price_data_ingestion]  (assertion, 0.99)  - Data Pipeline (service)
         ↓
[price_recommendation]  (suggestion, 0.75) - Pricing AI (llm)
         ↓
[price_decision]        (decision, 0.92)   - Sarah Chen (human)
         ↓
[price_execution]       (execution, 1.0)   - Pricing Engine (service)
```

## Code Highlights

### Initialize SDK

```typescript
import * as lineage from 'lineage-sdk';

await lineage.init({
  project: 'pricing-demo-ts',
  domain: 'pricing',
  baseUrl: 'http://localhost:8080',
  actorName: 'Pricing Service',
  actorType: 'service',
});
```

### Emit Events with Lineage

```typescript
// Data ingestion (no parent)
const dataEvent = await lineage.emit('price_data_ingestion', 'assertion', {
  productId: 'SKU-123',
  currentPrice: 29.99,
}, {
  confidence: 0.99,
  actor: ['service', 'Data Pipeline'],
});

// AI suggestion (links to data event)
const suggestionEvent = await lineage.emit('price_recommendation', 'suggestion', {
  recommendedPrice: 28.49,
  reasoning: 'Market analysis',
}, {
  confidence: 0.75,
  actor: ['llm', 'Pricing AI'],
  parent: dataEvent, // Link to parent
});

// Human decision (links to suggestion)
const decisionEvent = await lineage.emit('price_decision', 'decision', {
  approvedPrice: 27.99,
  reviewerNotes: 'Adjusted to psychological price point',
}, {
  confidence: 0.92,
  actor: ['human', 'Sarah Chen'],
  parent: suggestionEvent,
  reason: 'Adjusted to psychological price point',
});
```

## Sample Output

```
============================================================
Pricing Decision Workflow with Lineage (TypeScript)
============================================================

[Step 0] Initializing Lineage...
  Lineage initialized

[Step 1] Data Ingestion (assertion)
----------------------------------------
  Ingested data for SKU-123
  Current price: $29.99

[Step 2] AI Recommendation (suggestion)
----------------------------------------
  Recommended: $28.49
  Reasoning: Price is 10%+ above competitor average. Recommend 5% reduction.
  Confidence: 75%

[Step 3] Human Decision (decision)
----------------------------------------
  Approved: $27.99
  Modified: Yes
  Notes: Adjusted to psychological price point ($27.99)

[Step 4] Execution (execution)
----------------------------------------
  Price changed: $29.99 → $27.99
  Change: -6.7%

============================================================
Workflow Complete - Lineage Tracked
============================================================
```
