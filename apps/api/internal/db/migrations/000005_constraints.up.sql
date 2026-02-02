-- Constraints and triggers for append-only enforcement and data integrity

-- ============================================================
-- 1. APPEND-ONLY RULES
-- ============================================================
-- Applied to: events, event_scores, event_lineage, event_artifacts
-- UPDATE and DELETE are denied via DB rules.
-- Why rules? They protect against application bugs.

CREATE RULE events_no_update AS ON UPDATE TO events DO INSTEAD NOTHING;
CREATE RULE events_no_delete AS ON DELETE TO events DO INSTEAD NOTHING;

CREATE RULE event_scores_no_update AS ON UPDATE TO event_scores DO INSTEAD NOTHING;
CREATE RULE event_scores_no_delete AS ON DELETE TO event_scores DO INSTEAD NOTHING;

CREATE RULE event_lineage_no_update AS ON UPDATE TO event_lineage DO INSTEAD NOTHING;
CREATE RULE event_lineage_no_delete AS ON DELETE TO event_lineage DO INSTEAD NOTHING;

CREATE RULE event_artifacts_no_update AS ON UPDATE TO event_artifacts DO INSTEAD NOTHING;
CREATE RULE event_artifacts_no_delete AS ON DELETE TO event_artifacts DO INSTEAD NOTHING;

-- ============================================================
-- 2. CORRECTION INTEGRITY
-- ============================================================
-- Applied to: events

-- a) Both-or-neither CHECK: correction_type and corrects_event_id must both be set or both be null
ALTER TABLE events ADD CONSTRAINT chk_correction_pair
    CHECK (
        (correction_type IS NULL AND corrects_event_id IS NULL)
        OR
        (correction_type IS NOT NULL AND corrects_event_id IS NOT NULL)
    );

-- b) Same-scope + temporal ordering TRIGGER
CREATE OR REPLACE FUNCTION trg_correction_integrity()
    RETURNS TRIGGER AS $$
BEGIN
    IF NEW.corrects_event_id IS NOT NULL THEN
        -- Corrected event must be in the same scope
        IF NOT EXISTS (
            SELECT 1 FROM events
            WHERE id = NEW.corrects_event_id
              AND scope_id = NEW.scope_id
        ) THEN
            RAISE EXCEPTION 'Corrected event must be in the same scope';
        END IF;

        -- Corrected event must precede this one
        IF NOT EXISTS (
            SELECT 1 FROM events
            WHERE id = NEW.corrects_event_id
              AND scope_sequence < NEW.scope_sequence
        ) THEN
            RAISE EXCEPTION 'Corrected event must precede correcting event';
        END IF;
    END IF;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER enforce_correction_integrity
    BEFORE INSERT ON events
    FOR EACH ROW EXECUTE FUNCTION trg_correction_integrity();

-- ============================================================
-- 3. SCORE RANGE
-- ============================================================
-- Applied to: event_scores
-- Value must be between 0.0 and 1.0

ALTER TABLE event_scores ADD CONSTRAINT chk_score_range
    CHECK (value >= 0.0 AND value <= 1.0);
