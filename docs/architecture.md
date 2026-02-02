# Architecture

## Overview

Lineage is an append-only event store with hash-chaining for tracking AI decisions and their epistemic context.

## Components

```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│   Client    │────▶│   API       │────▶│   Kafka     │
│             │     │  (Gin)      │     │             │
└─────────────┘     └─────────────┘     └──────┬──────┘
                                               │
                                               ▼
                                        ┌─────────────┐
                                        │  Consumer   │
                                        │             │
                                        └──────┬──────┘
                                               │
                                               ▼
                                        ┌─────────────┐
                                        │  Postgres   │
                                        │             │
                                        └─────────────┘
```

### API Server (`apps/api/cmd/api`)

- REST API built with Gin
- Validates and accepts events
- Produces events to Kafka
- Serves Swagger documentation at `/swagger/*`

### Consumer (`apps/api/cmd/consumer`)

- Consumes events from Kafka
- Computes hash chains (RFC 8785 canonical JSON)
- Persists events to Postgres
- Maintains scope sequences

### Database (Postgres)

- Append-only event storage
- Hash chain verification
- Lineage relationships (parent/child events)

### Message Queue (Kafka)

- Asynchronous event processing
- SASL/SSL authentication (Aiven)
- Single topic: `lineage.events`

## Data Model

### Core Entities

- **Scope**: Logical grouping (project/domain/environment)
- **Actor**: Who performed the action (human, LLM, agent, service, tool)
- **Event Type**: Schema for event payloads
- **Event**: Immutable record with hash chain

### Hash Chaining

Each event includes:
- `event_hash`: SHA-256 of canonical JSON (RFC 8785)
- `prev_event_hash`: Links to previous event in scope
- `scope_sequence`: Monotonic sequence number within scope

## Directory Structure

```
Lineage/
├── apps/
│   └── api/                    # Go backend
│       ├── cmd/
│       │   ├── api/            # HTTP server
│       │   └── consumer/       # Kafka consumer
│       ├── docs/               # Swagger docs (generated)
│       └── internal/
│           ├── app/            # Infrastructure (config, db, kafka)
│           ├── actor/          # Actor feature
│           ├── event/          # Event feature
│           ├── eventtype/      # Event type feature
│           ├── health/         # Health checks
│           ├── lineage/        # Lineage relationships
│           └── scope/          # Scope feature
├── docs/                       # Project documentation
├── infra/                      # Infrastructure configs
├── schema.dbml                 # Database schema
└── ui/                         # Frontend (future)
```
