-- name: CreateActor :one
INSERT INTO actors (type, external_id, name, metadata)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: GetActor :one
SELECT * FROM actors WHERE id = $1;

-- name: GetActorByExternalID :one
SELECT * FROM actors WHERE type = $1 AND external_id = $2;

-- name: ListActors :many
SELECT * FROM actors ORDER BY registered_at DESC;

-- name: ListActorsByType :many
SELECT * FROM actors WHERE type = $1 ORDER BY registered_at DESC;
