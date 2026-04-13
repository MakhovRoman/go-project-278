-- name: GetListLinks :many
SELECT id, original_url, short_name, short_url, created_at
FROM links
ORDER BY id LIMIT $1 OFFSET $2;

-- name: CountLinks :one
SELECT COUNT(*) FROM links;


-- name: NewLink :one
INSERT INTO links (original_url, short_name, short_url)
VALUES ($1, $2, $3)
RETURNING id, original_url, short_name, short_url, created_at;

-- name: GetLinkByID :one
SELECT id, original_url, short_name, short_url, created_at
FROM links
WHERE id=$1;

-- name: UpdateLinkByID :one
UPDATE links
SET original_url=$2,
    short_name=$3,
    short_url=$4
WHERE id=$1
    RETURNING id, original_url, short_name, short_url, created_at;


-- name: DeleteLink :exec
DELETE FROM links
WHERE id=$1;
