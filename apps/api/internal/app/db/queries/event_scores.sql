-- name: InsertScore :one
INSERT INTO event_scores (event_id, type, value, category, scored_by, reason, metadata)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetScoresForEvent :many
SELECT * FROM event_scores WHERE event_id = $1 ORDER BY created_at DESC;

-- name: GetScoresByType :many
SELECT * FROM event_scores WHERE event_id = $1 AND type = $2 ORDER BY created_at DESC;

-- name: GetLatestScoreByType :one
SELECT * FROM event_scores
WHERE event_id = $1 AND type = $2
ORDER BY created_at DESC
LIMIT 1;
