package database

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
)

type DBStore struct {
	db *sql.DB
}

func NewDBStore(db *sql.DB) *DBStore {
	return &DBStore{db: db}
}

func (s *DBStore) Save(shortURL, originalURL string) error {
	_, err := s.db.Exec(`
		INSERT INTO urls (short_url, original_url) 
		VALUES ($1, $2)
		ON CONFLICT (original_url) DO NOTHING
	`, shortURL, originalURL)
	return err
}

func (s *DBStore) Get(shortURL string) (string, error) {
	var originalURL string
	err := s.db.QueryRow("SELECT original_url FROM urls WHERE short_url=$1", shortURL).Scan(&originalURL)
	if err != nil {
		return "", err
	}
	return originalURL, nil
}

func (s *DBStore) GetByOriginalURL(originalURL string) (string, error) {
	var shortURL string
	err := s.db.QueryRow("SELECT short_url FROM urls WHERE original_url=$1", originalURL).Scan(&shortURL)
	if err != nil {
		return "", err
	}
	return shortURL, nil
}

func (s *DBStore) SaveWithConflict(shortURL, originalURL string) (string, error) {
	_, err := s.db.Exec(`
		INSERT INTO urls (short_url, original_url) 
		VALUES ($1, $2)
		ON CONFLICT (original_url) DO NOTHING
	`, shortURL, originalURL)

	if err == nil {
		return shortURL, nil // Успешно вставили новую запись
	}

	// Проверяем тип ошибки
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == pgerrcode.UniqueViolation {
		// Получаем уже существующий shortURL из базы
		existingShortURL, err2 := s.GetByOriginalURL(originalURL)
		if err2 != nil {
			return "", err2
		}
		return existingShortURL, nil
	}

	return "", err
}
