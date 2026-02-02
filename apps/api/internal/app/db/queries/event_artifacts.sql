-- name: LinkArtifact :one
INSERT INTO event_artifacts (event_id, artifact_id, role)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetArtifactsForEvent :many
SELECT a.* FROM artifacts a
JOIN event_artifacts ea ON ea.artifact_id = a.id
WHERE ea.event_id = $1;

-- name: GetInputArtifactsForEvent :many
SELECT a.* FROM artifacts a
JOIN event_artifacts ea ON ea.artifact_id = a.id
WHERE ea.event_id = $1 AND ea.role = 'input';

-- name: GetOutputArtifactsForEvent :many
SELECT a.* FROM artifacts a
JOIN event_artifacts ea ON ea.artifact_id = a.id
WHERE ea.event_id = $1 AND ea.role = 'output';

-- name: GetEventsForArtifact :many
SELECT e.* FROM events e
JOIN event_artifacts ea ON ea.event_id = e.id
WHERE ea.artifact_id = $1;
