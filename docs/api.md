# API Reference

## Interactive Documentation

When the API is running, visit the Swagger UI:

```
http://localhost:8080/swagger/index.html
```

## Base URL

```
http://localhost:8080/api/v1
```

## Endpoints

### Health

| Method | Path | Description |
|--------|------|-------------|
| GET | `/health` | Check API and dependency health |

### Scopes

| Method | Path | Description |
|--------|------|-------------|
| POST | `/scopes` | Create a new scope |
| GET | `/scopes` | List all scopes |
| GET | `/scopes/:id` | Get scope by ID |

### Actors

| Method | Path | Description |
|--------|------|-------------|
| POST | `/actors` | Create a new actor |
| GET | `/actors` | List all actors |
| GET | `/actors/:id` | Get actor by ID |

### Event Types

| Method | Path | Description |
|--------|------|-------------|
| POST | `/event-types` | Create a new event type |
| GET | `/event-types` | List all active event types |
| GET | `/event-types/:id` | Get event type by ID |

### Events

| Method | Path | Description |
|--------|------|-------------|
| POST | `/events` | Submit event to processing queue |
| GET | `/events?scope_id=` | List events by scope |
| GET | `/events/:id` | Get event by ID |
| GET | `/events/:id/lineage` | Get parent/child events |

### Scores

| Method | Path | Description |
|--------|------|-------------|
| POST | `/events/:id/scores` | Add score to event |
| GET | `/events/:id/scores` | Get all scores for event |
| GET | `/events/:id/scores?type=confidence` | Get scores filtered by type |

### Artifacts

| Method | Path | Description |
|--------|------|-------------|
| POST | `/artifacts` | Create artifact |
| GET | `/artifacts/:id` | Get artifact by ID |
| GET | `/artifacts?scope_id=` | List artifacts in scope |
| GET | `/artifacts?scope_id=&content_hash=` | Find artifact by hash (dedup) |
| POST | `/events/:id/artifacts` | Link artifact to event |
| GET | `/events/:id/artifacts` | Get artifacts for event |

## Event Intents

Valid values for `intent` field:
- `exploration` - Gathering information
- `suggestion` - Proposing an option
- `assertion` - Stating a fact
- `decision` - Making a choice
- `execution` - Taking an action

## Correction Types

Valid values for `correction_type` field:
- `supersede` - Replaces previous event
- `amend` - Adds to previous event
- `retract` - Withdraws previous event

## Score Types

Valid values for score `type` field:
- `confidence` - How certain the actor is
- `relevance` - How relevant to the scope
- `reliability` - How reliable the source/data is
- `agreement` - Level of agreement with prior assertions

Score categories (auto-derived from value):
| Value | Category |
|-------|----------|
| 0.00-0.19 | very_low |
| 0.20-0.39 | low |
| 0.40-0.59 | moderate |
| 0.60-0.79 | high |
| 0.80-1.00 | very_high |

## Artifact Roles

Valid values for artifact `role` field:
- `input` - Data consumed by the event
- `output` - Data produced by the event

## Inline Event Fields

Events support inline confidence and artifact linking:

```json
{
  "scope_id": "...",
  "actor_id": "...",
  "event_type_id": "...",
  "intent": "suggestion",
  "confidence": 0.85,
  "input_artifact_ids": ["..."],
  "output_artifact_ids": ["..."],
  "payload": {}
}
```

## OpenAPI Spec

Raw OpenAPI files are available at:
- JSON: `apps/api/docs/swagger.json`
- YAML: `apps/api/docs/swagger.yaml`
