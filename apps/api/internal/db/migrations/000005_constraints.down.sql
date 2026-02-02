-- Remove constraints and triggers

-- Score range
ALTER TABLE event_scores DROP CONSTRAINT IF EXISTS chk_score_range;

-- Correction integrity
DROP TRIGGER IF EXISTS enforce_correction_integrity ON events;
DROP FUNCTION IF EXISTS trg_correction_integrity();
ALTER TABLE events DROP CONSTRAINT IF EXISTS chk_correction_pair;

-- Append-only rules
DROP RULE IF EXISTS event_artifacts_no_delete ON event_artifacts;
DROP RULE IF EXISTS event_artifacts_no_update ON event_artifacts;
DROP RULE IF EXISTS event_lineage_no_delete ON event_lineage;
DROP RULE IF EXISTS event_lineage_no_update ON event_lineage;
DROP RULE IF EXISTS event_scores_no_delete ON event_scores;
DROP RULE IF EXISTS event_scores_no_update ON event_scores;
DROP RULE IF EXISTS events_no_delete ON events;
DROP RULE IF EXISTS events_no_update ON events;
