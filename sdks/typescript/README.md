# Lineage TypeScript SDK

Epistemic transparency for AI systems - track decisions, suggestions, and actions with full lineage.

## Installation

```bash
npm install lineage-sdk
# or
pnpm add lineage-sdk
# or
yarn add lineage-sdk
```

## Quick Start

```typescript
import * as lineage from 'lineage-sdk';

// Initialize once
await lineage.init({
  project: 'my-app',
  actorName: 'my-service',
  actorType: 'service'
});

// Emit events
await lineage.emit('data_ingestion', 'assertion', { data: 'loaded' }, { confidence: 0.99 });

// Or use track wrapper
const recommend = lineage.track('recommendation', 'suggestion', {
  actor: ['llm', 'GPT-4']
})(async (data) => {
  return { price: 26.99, confidence: 0.85 };
});
```

## Simple API

### Initialize

```typescript
import * as lineage from 'lineage-sdk';

await lineage.init({
  project: 'my-app',                  // Required: project name
  baseUrl: 'http://localhost:8080',   // Lineage API server
  domain: 'pricing',                  // Optional: domain
  environment: 'production',          // Optional: environment
  actorName: 'my-service',            // Default actor name
  actorType: 'service',               // human, llm, agent, service, tool
});
```

### Emit Events

```typescript
await lineage.emit(
  'event_type_name',      // Auto-created if needed
  'assertion',            // intent
  { key: 'value' },       // payload
  {
    confidence: 0.95,             // optional
    actor: ['llm', 'GPT-4'],      // optional override
    parent: previousEvent,         // optional parent
  }
);
```

### Track Functions

```typescript
const recommend = lineage.track('recommendation', 'suggestion', {
  actor: ['llm', 'Pricing AI'],
  confidence: 0.72
})(async (data: Input) => {
  // Return value becomes payload
  // 'confidence' key is auto-extracted
  return {
    recommendedPrice: 26.99,
    reasoning: 'Market analysis',
    confidence: 0.72
  };
});

const result = await recommend(inputData); // Event auto-emitted
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
| `agent` | Autonomous agents |
| `service` | Backend services, APIs |
| `tool` | Tools, functions |

## Full Example

```typescript
import * as lineage from 'lineage-sdk';

await lineage.init({ project: 'ecommerce', domain: 'pricing' });

// Service ingests data
const e1 = await lineage.emit(
  'data_ingestion', 'assertion',
  { product: 'SKU-123', price: 29.99 },
  { confidence: 0.99, actor: ['service', 'Data Pipeline'] }
);

// AI recommends
const recommend = lineage.track('recommendation', 'suggestion', {
  actor: ['llm', 'Pricing AI']
})(async (data) => {
  return { recommendedPrice: 26.99, confidence: 0.72 };
});

const rec = await recommend({ product: 'SKU-123' });

// Human decides
const e3 = await lineage.emit(
  'recommendation', 'decision',
  { approvedPrice: 27.99 },
  { confidence: 0.88, actor: ['human', 'Manager'], parent: lineage.getLastEvent() }
);

// Service executes
await lineage.emit(
  'execution', 'execution',
  { newPrice: 27.99 },
  { confidence: 1.0, actor: ['service', 'Pricing Engine'], parent: e3 }
);
```

## Low-Level API

For advanced use cases:

```typescript
import { LineageClient, type ActorType, type Intent } from 'lineage-sdk';

const client = new LineageClient({ baseUrl: 'http://localhost:8080' });

// Full control over resources
const scope = await client.scopes.create({ project: 'my-project' });
const actor = await client.actors.create({ type: 'llm', name: 'My LLM' });
const eventType = await client.eventTypes.create({
  name: 'custom_event',
  version: '1.0',
  allowed_intents: ['suggestion']
});

await client.events.create({
  scope_id: scope.id,
  actor_id: actor.id,
  event_type_id: eventType.id,
  intent: 'suggestion',
  payload: { data: 'value' }
});

// Query lineage
const lineage = await client.events.getLineage(eventId);
console.log(lineage.parents, lineage.children);
```

## Error Handling

```typescript
import { LineageError, NotFoundError, ValidationError, ServerError } from 'lineage-sdk';

try {
  const event = await client.events.get('non-existent-id');
} catch (error) {
  if (error instanceof NotFoundError) {
    console.log('Not found:', error.message);
  } else if (error instanceof ValidationError) {
    console.log('Invalid request:', error.message);
  } else if (error instanceof ServerError) {
    console.log('Server error:', error.message);
  }
}
```

## Development

```bash
npm install
npm run build
npm test
```

## License

Elastic-2.0
