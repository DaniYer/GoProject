// Code generated by sqlc. DO NOT EDIT.
// versions:
//   sqlc v1.29.0
// source: select_by_short.sql

package queries

import (
	"context"
)

const getByShortURL = `-- name: GetByShortURL :one
SELECT original_url FROM urls WHERE short_url = $1
`

func (q *Queries) GetByShortURL(ctx context.Context, shortUrl string) (string, error) {
	row := q.db.QueryRowContext(ctx, getByShortURL, shortUrl)
	var original_url string
	err := row.Scan(&original_url)
	return original_url, err
}
