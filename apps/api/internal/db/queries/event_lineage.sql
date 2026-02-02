-- name: CreateLineage :one
INSERT INTO event_lineage (parent_event_id, child_event_id, relationship)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetParents :many
SELECT e.* FROM events e
JOIN event_lineage el ON el.parent_event_id = e.id
WHERE el.child_event_id = $1;

-- name: GetChildren :many
SELECT e.* FROM events e
JOIN event_lineage el ON el.child_event_id = e.id
WHERE el.parent_event_id = $1;

-- name: GetLineageEdge :one
SELECT * FROM event_lineage
WHERE parent_event_id = $1 AND child_event_id = $2;

-- name: GetDirectParentIDs :many
SELECT parent_event_id FROM event_lineage WHERE child_event_id = $1;

-- name: GetDirectChildIDs :many
SELECT child_event_id FROM event_lineage WHERE parent_event_id = $1;
