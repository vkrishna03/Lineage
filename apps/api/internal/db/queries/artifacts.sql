-- name: CreateArtifact :one
INSERT INTO artifacts (scope_id, content_hash, content_type, uri, metadata)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetArtifact :one
SELECT * FROM artifacts WHERE id = $1;

-- name: GetArtifactByHash :one
SELECT * FROM artifacts WHERE scope_id = $1 AND content_hash = $2;

-- name: ListArtifactsInScope :many
SELECT * FROM artifacts WHERE scope_id = $1 ORDER BY created_at DESC LIMIT $2 OFFSET $3;
