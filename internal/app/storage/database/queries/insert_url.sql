-- name: InsertOrGetShortURL :one
INSERT INTO urls (short_url, original_url, user_id, is_deleted)
VALUES ($1, $2, $3, false)
ON CONFLICT (original_url) DO NOTHING
RETURNING short_url;
