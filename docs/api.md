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

## OpenAPI Spec

Raw OpenAPI files are available at:
- JSON: `apps/api/docs/swagger.json`
- YAML: `apps/api/docs/swagger.yaml`
