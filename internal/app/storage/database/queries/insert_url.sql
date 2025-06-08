-- name: InsertOrGetShortURL :one
INSERT INTO urls (short_url, original_url, user_id)
VALUES ($1, $2, $3)
ON CONFLICT (original_url) DO NOTHING
RETURNING short_url;