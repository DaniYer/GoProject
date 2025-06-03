package database

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// DBStore – реализация URLStore через PostgreSQL.
type DBStore struct {
	db *sql.DB
}

func NewDBStore(db *sql.DB) *DBStore {
	return &DBStore{db: db}
}

func (s *DBStore) Save(shortURL, originalURL string) error {
	query := "INSERT INTO urls (short_url, original_url) VALUES ($1, $2)"
	_, err := s.db.Exec(query, shortURL, originalURL)
	return err
}

func (s *DBStore) Get(shortURL string) (string, error) {
	var originalURL string
	query := "SELECT original_url FROM urls WHERE short_url=$1"
	err := s.db.QueryRow(query, shortURL).Scan(&originalURL)
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

// initDB устанавливает соединение с базой данных.
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

func CreateTable(db *sql.DB) error {
	query := `
	CREATE TABLE IF NOT EXISTS urls (
		id SERIAL PRIMARY KEY,
		short_url VARCHAR(8) UNIQUE NOT NULL,
		original_url TEXT NOT NULL
	);`
	_, err := db.Exec(query)
	return err
}
