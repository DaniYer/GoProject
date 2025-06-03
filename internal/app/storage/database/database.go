package database

import (
	"context"
	"database/sql"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// DBStore – реализация URLStore через PostgreSQL.
type DBStore struct {
	db *sql.DB
}

func NewDBStore(db *sql.DB) *DBStore {
	return &DBStore{db: db}
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
