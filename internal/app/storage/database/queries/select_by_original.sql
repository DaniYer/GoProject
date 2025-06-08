-- name: GetByOriginalURL :one
SELECT short_url FROM urls WHERE original_url = $1;
