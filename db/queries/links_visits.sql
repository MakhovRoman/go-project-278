-- name: CreateVisit :one
INSERT INTO link_visits(link_id, ip, user_agent, referer, status)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, link_id, created_at, ip, user_agent, referer, status;

-- name: GetListVisits :many
SELECT id, link_id, created_at, ip, user_agent, referer, status
FROM link_visits
ORDER BY id LIMIT $1 OFFSET $2;

-- name: CountVisits :one
SELECT COUNT(*)
FROM link_visits;
