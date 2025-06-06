package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/DaniYer/GoProject.git/internal/app/storage/database/queries"
	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

type DBStore struct {
	db      *sql.DB
	queries *queries.Queries
}

func NewDBStore(db *sql.DB) *DBStore {
	return &DBStore{
		db:      db,
		queries: queries.New(db),
	}
}
func (s *DBStore) Save(shortURL, originalURL string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	newShortURL, err := s.queries.InsertOrGetShortURL(ctx, queries.InsertOrGetShortURLParams{
		ShortUrl:    shortURL,
		OriginalUrl: originalURL,
	})

	if errors.Is(err, pgx.ErrNoRows) {
		existingShortURL, err2 := s.queries.GetByOriginalURL(ctx, originalURL)
		if err2 != nil {
			return "", err2
		}
		return existingShortURL, nil
	}

	if err != nil {
		return "", err
	}

	return newShortURL, nil
}

func (s *DBStore) Get(shortURL string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	result, err := s.queries.GetByShortURL(ctx, shortURL)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (s *DBStore) GetByOriginalURL(originalURL string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	result, err := s.queries.GetByOriginalURL(ctx, originalURL)
	if err != nil {
		return "", err
	}
	return result, nil
}

func (s *DBStore) SaveWithConflict(shortURL, originalURL string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	shortURL, err := s.queries.InsertOrGetShortURL(ctx, queries.InsertOrGetShortURLParams{
		ShortUrl:    shortURL,
		OriginalUrl: originalURL,
	})
	if err == nil {
		return shortURL, nil
	}

	// Проверяем тип ошибки
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		existingShortURL, err2 := s.GetByOriginalURL(originalURL)
		if err2 != nil {
			return "", err2
		}
		return existingShortURL, nil
	}

	return "", err
}

func InitDB(driverName, dataSourceName string) (*sql.DB, error) {
	db, err := sql.Open(driverName, dataSourceName)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		db.Close()
		return nil, err
	}
	return db, nil
}
