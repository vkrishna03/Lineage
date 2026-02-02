-- name: InsertEvent :one
INSERT INTO events (
    scope_id, actor_id, event_type_id, scope_sequence, intent, reason,
    correction_type, corrects_event_id, observed_at, decided_at,
    prev_event_hash, event_hash, payload
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
RETURNING *;

-- name: GetEvent :one
SELECT * FROM events WHERE id = $1;

-- name: GetEventByHash :one
SELECT * FROM events WHERE event_hash = $1;

-- name: GetEventsByScope :many
SELECT * FROM events
WHERE scope_id = $1
ORDER BY scope_sequence ASC
LIMIT $2 OFFSET $3;

-- name: GetLastEventInScope :one
SELECT * FROM events
WHERE scope_id = $1
ORDER BY scope_sequence DESC
LIMIT 1;

-- name: GetEventsByActor :many
SELECT * FROM events
WHERE actor_id = $1
ORDER BY ingested_at DESC
LIMIT $2 OFFSET $3;

-- name: GetCorrectionsForEvent :many
SELECT * FROM events
WHERE corrects_event_id = $1
ORDER BY scope_sequence ASC;

-- name: CountEventsInScope :one
SELECT COUNT(*) FROM events WHERE scope_id = $1;
