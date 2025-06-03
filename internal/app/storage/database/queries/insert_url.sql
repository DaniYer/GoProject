-- name: InsertOrGetShortURL :one
INSERT INTO urls (short_url, original_url)
VALUES ($1, $2)
ON CONFLICT (original_url) DO NOTHING
RETURNING short_url;
