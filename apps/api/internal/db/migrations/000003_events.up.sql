-- Events table (append-only)
-- Immutable records of actions. Never updated or deleted.

CREATE TABLE events (
    id                  UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scope_id            UUID NOT NULL REFERENCES scopes(id),
    actor_id            UUID NOT NULL REFERENCES actors(id),
    event_type_id       UUID NOT NULL REFERENCES event_type_registry(id),

    -- Ordering: derived from Kafka partition offset
    -- Guarantees total order within a scope
    scope_sequence      BIGINT NOT NULL,

    -- Intent: every event MUST declare its epistemic posture
    intent              event_intent NOT NULL,

    -- Human-readable explanation of WHY this event was emitted
    reason              TEXT,

    -- Correction chain: corrections reference prior events
    correction_type     correction_type,
    corrects_event_id   UUID REFERENCES events(id),

    -- Time semantics
    observed_at         TIMESTAMPTZ,  -- when the action happened in real world
    decided_at          TIMESTAMPTZ,  -- when a commitment was made
    ingested_at         TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    -- Hash chain for append-only integrity
    prev_event_hash     VARCHAR,  -- NULL only for genesis event (scope_sequence = 0)
    event_hash          VARCHAR NOT NULL,

    -- Payload: structure governed by event_type_registry.payload_schema
    payload             JSONB NOT NULL
);

-- Indexes for events table
CREATE UNIQUE INDEX idx_events_scope_sequence ON events(scope_id, scope_sequence);
CREATE INDEX idx_events_scope ON events(scope_id);
CREATE INDEX idx_events_actor ON events(actor_id);
CREATE INDEX idx_events_type ON events(event_type_id);
CREATE INDEX idx_events_corrects ON events(corrects_event_id);
CREATE INDEX idx_events_ingested ON events(ingested_at);
CREATE UNIQUE INDEX idx_events_hash ON events(event_hash);

COMMENT ON TABLE events IS 'IMMUTABLE. Never updated or deleted. See append-only enforcement.';
