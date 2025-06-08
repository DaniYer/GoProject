-- name: GetAllByUserID :many
SELECT short_url, original_url FROM urls WHERE user_id = $1;