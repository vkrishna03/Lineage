-- Supporting tables: event_scores, artifacts, event_artifacts, event_lineage

-- Event Scores: normalized scores separated from events
-- Multiple score types coexist, multiple actors can score the same event,
-- scores can arrive after the event (async assessment)
CREATE TABLE event_scores (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id    UUID NOT NULL REFERENCES events(id),
    type        score_type NOT NULL,
    value       DECIMAL NOT NULL,  -- 0.0–1.0 normalized. Source of truth.
    category    score_category NOT NULL,  -- Derived from value at app layer
    scored_by   UUID REFERENCES actors(id),  -- NULL = self-reported by event actor
    reason      TEXT,
    metadata    JSONB,  -- Model name, scoring method, thresholds, etc.
    created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_event_scores_event ON event_scores(event_id);
CREATE UNIQUE INDEX idx_event_scores_unique ON event_scores(event_id, type, scored_by);

COMMENT ON TABLE event_scores IS 'Append-only. A new assessment is a new row, not an update.';

-- Artifacts: data objects produced or consumed by events
-- Content-addressed via content_hash for deduplication within a scope.
-- Actual binary/text content lives in external object storage.
CREATE TABLE artifacts (
    id            UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    scope_id      UUID NOT NULL REFERENCES scopes(id),
    content_hash  VARCHAR NOT NULL,  -- SHA-256 of artifact content
    content_type  VARCHAR NOT NULL,  -- MIME type
    uri           VARCHAR,           -- External storage reference (S3, GCS, etc.)
    metadata      JSONB,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_artifacts_scope_hash ON artifacts(scope_id, content_hash);

-- Event Artifacts: junction table for event ↔ artifact relationships
CREATE TABLE event_artifacts (
    id          UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id    UUID NOT NULL REFERENCES events(id),
    artifact_id UUID NOT NULL REFERENCES artifacts(id),
    role        artifact_role NOT NULL  -- input = consumed, output = produced
);

CREATE UNIQUE INDEX idx_event_artifacts_unique ON event_artifacts(event_id, artifact_id, role);
CREATE INDEX idx_event_artifacts_artifact ON event_artifacts(artifact_id);

-- Event Lineage: explicit causal edges between events (DAG)
-- This is the backbone of "where did this come from?" queries
CREATE TABLE event_lineage (
    id              UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_event_id UUID NOT NULL REFERENCES events(id),
    child_event_id  UUID NOT NULL REFERENCES events(id),
    relationship    VARCHAR NOT NULL  -- caused_by | derived_from | triggered_by | informed_by
);

CREATE INDEX idx_event_lineage_parent ON event_lineage(parent_event_id);
CREATE INDEX idx_event_lineage_child ON event_lineage(child_event_id);
CREATE UNIQUE INDEX idx_event_lineage_unique ON event_lineage(parent_event_id, child_event_id, relationship);
