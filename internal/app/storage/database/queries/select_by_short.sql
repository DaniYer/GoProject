-- name: GetByShortURL :one
SELECT original_url FROM urls WHERE short_url = $1;
