-- Core tables: scopes, actors, event_type_registry

-- Scopes: Trust boundaries
-- Every event and artifact lives inside a scope.
-- Scopes prevent cross-project, cross-environment trust leakage.
CREATE TABLE scopes (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    project     VARCHAR NOT NULL,
    domain      VARCHAR,
    environment VARCHAR,  -- e.g. production | staging | sandbox
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (project, domain, environment)
);

-- Actors: anything that can emit an event
-- The system does not own identity — external_id references
-- the actor in the originating system (IAM, model registry, etc.)
CREATE TABLE actors (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type          actor_type NOT NULL,
    external_id   VARCHAR,  -- Identifier in the originating system
    name          VARCHAR,
    metadata      JSONB,
    registered_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (type, external_id)
);

-- Event Type Registry: protocol catalog
-- Every event must reference a registered type+version.
-- Defines what event shapes are legal and what their payloads look like.
CREATE TABLE event_type_registry (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name            VARCHAR NOT NULL,         -- e.g. recommendation.created, data.ingested
    version         VARCHAR NOT NULL,         -- e.g. v1, v2
    description     TEXT,
    payload_schema  JSONB,                    -- JSON Schema defining valid shape of event.payload
    allowed_intents VARCHAR[],                -- Subset of event_intent valid for this type
    is_active       BOOLEAN NOT NULL DEFAULT TRUE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE (name, version)
);

COMMENT ON TABLE event_type_registry IS 'Mutable table — this is configuration, not event data. New versions of an event type get new rows; old rows are deactivated (is_active = false), never deleted.';
