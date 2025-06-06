-- +goose Up
CREATE TABLE IF NOT EXISTS urls (
    id SERIAL PRIMARY KEY,
    short_url TEXT NOT NULL UNIQUE,
    original_url TEXT NOT NULL UNIQUE, 
    user_id TEXT NOT NULL,
    is_deleted BOOLEAN DEFAULT FALSE
);

-- +goose Down
DROP TABLE IF EXISTS urls;
