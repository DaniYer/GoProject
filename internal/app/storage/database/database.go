package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/DaniYer/GoProject.git/internal/app/dto"
	"github.com/DaniYer/GoProject.git/internal/app/storage/database/queries"
	"github.com/jackc/pgx"
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

func (s *DBStore) Save(shortURL, originalURL, userID string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	newShortURL, err := s.queries.InsertOrGetShortURL(ctx, queries.InsertOrGetShortURLParams{
		ShortUrl:    shortURL,
		OriginalUrl: originalURL,
		UserID:      sql.NullString{String: userID, Valid: true},
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

func (s *DBStore) GetAllByUser(userID string) ([]dto.UserURL, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	urls, err := s.queries.GetAllByUserID(ctx, sql.NullString{String: userID, Valid: true})
	if err != nil {
		return nil, err
	}

	result := make([]dto.UserURL, 0, len(urls))
	for _, u := range urls {
		result = append(result, dto.UserURL{
			ShortURL:    u.ShortUrl,
			OriginalURL: u.OriginalUrl,
		})
	}
	return result, nil
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
