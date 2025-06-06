package repository

import (
	"context"
	"errors"
	"time"

	"github.com/DaniYer/GoProject.git/internal/app/dto"
	"github.com/google/uuid"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

type PostgresStorage struct {
	db *pgx.Conn
}

var ErrDuplicateURL = errors.New("duplicate url")

func NewPostgres(dsn string) (*PostgresStorage, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		return nil, err
	}
	return &PostgresStorage{db: conn}, nil
}

func (p *PostgresStorage) Ping(ctx context.Context) error {
	return p.db.Ping(ctx)
}

func (p *PostgresStorage) Close(ctx context.Context) {
	_ = p.db.Close(ctx)
}

func (p *PostgresStorage) InitSchema(ctx context.Context) error {
	query := `
	CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		short_url TEXT NOT NULL UNIQUE,
		original_url TEXT NOT NULL UNIQUE,
		user_id TEXT NOT NULL,
		is_deleted BOOLEAN DEFAULT FALSE
	)`
	_, err := p.db.Exec(ctx, query)
	return err
}

func (p *PostgresStorage) SaveURL(ctx context.Context, userID, originalURL string) (string, error) {
	shortID := uuid.New().String()[:8]

	_, err := p.db.Exec(ctx,
		`INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)`,
		shortID, originalURL, userID)

	if err == nil {
		return shortID, nil
	}

	if pgerr, ok := err.(*pgconn.PgError); ok {
		if pgerr.Code == pgerrcode.UniqueViolation {
			var existingShortID string
			query := `SELECT short_url FROM urls WHERE original_url = $1 LIMIT 1`
			err2 := p.db.QueryRow(ctx, query, originalURL).Scan(&existingShortID)
			if err2 != nil {
				return "", err2
			}
			return existingShortID, ErrDuplicateURL
		}
	}

	return "", err
}

func (p *PostgresStorage) GetOriginalURL(ctx context.Context, shortID string) (string, error) {
	var url string
	err := p.db.QueryRow(ctx,
		`SELECT original_url FROM urls WHERE short_url = $1 AND is_deleted = FALSE`,
		shortID).Scan(&url)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", nil
		}
		return "", err
	}
	return url, nil
}

func (p *PostgresStorage) GetUserURLs(ctx context.Context, userID string) ([]dto.UserURL, error) {
	rows, err := p.db.Query(ctx,
		`SELECT short_url, original_url FROM urls WHERE user_id = $1 AND is_deleted = FALSE`,
		userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []dto.UserURL
	for rows.Next() {
		var u dto.UserURL
		if err := rows.Scan(&u.ShortURL, &u.OriginalURL); err != nil {
			return nil, err
		}
		result = append(result, u)
	}
	return result, nil
}

func (p *PostgresStorage) SaveBatchURLs(ctx context.Context, userID string, batch []dto.BatchRequestItem, baseURL string) ([]dto.BatchResponseItem, error) {
	tx, err := p.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var result []dto.BatchResponseItem

	for _, item := range batch {
		shortID := uuid.New().String()[:8]
		_, err := tx.Exec(ctx,
			`INSERT INTO urls (short_url, original_url, user_id) VALUES ($1, $2, $3)`,
			shortID, item.OriginalURL, userID)

		if err != nil {
			if pgerr, ok := err.(*pgconn.PgError); ok {
				if pgerr.Code == pgerrcode.UniqueViolation {
					var existingShortID string
					query := `SELECT short_url FROM urls WHERE original_url = $1 LIMIT 1`
					err2 := tx.QueryRow(ctx, query, item.OriginalURL).Scan(&existingShortID)
					if err2 != nil {
						return nil, err2
					}
					result = append(result, dto.BatchResponseItem{
						CorrelationID: item.CorrelationID,
						ShortURL:      baseURL + "/" + existingShortID,
					})
					continue
				}
			}
			return nil, err
		}

		result = append(result, dto.BatchResponseItem{
			CorrelationID: item.CorrelationID,
			ShortURL:      baseURL + "/" + shortID,
		})
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return result, nil
}
