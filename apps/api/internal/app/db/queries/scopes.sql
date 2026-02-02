-- name: CreateScope :one
INSERT INTO scopes (project, domain, environment)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetScope :one
SELECT * FROM scopes WHERE id = $1;

-- name: GetScopeByProject :one
SELECT * FROM scopes
WHERE project = $1
  AND (domain = $2 OR (domain IS NULL AND $2 IS NULL))
  AND (environment = $3 OR (environment IS NULL AND $3 IS NULL));

-- name: ListScopes :many
SELECT * FROM scopes ORDER BY created_at DESC;
