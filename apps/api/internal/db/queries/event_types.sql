-- name: CreateEventType :one
INSERT INTO event_type_registry (name, version, description, payload_schema, allowed_intents)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetEventType :one
SELECT * FROM event_type_registry WHERE id = $1;

-- name: GetEventTypeByNameVersion :one
SELECT * FROM event_type_registry WHERE name = $1 AND version = $2;

-- name: ListActiveEventTypes :many
SELECT * FROM event_type_registry WHERE is_active = true ORDER BY name, version;

-- name: DeactivateEventType :exec
UPDATE event_type_registry SET is_active = false WHERE id = $1;
