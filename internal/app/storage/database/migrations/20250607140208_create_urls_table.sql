-- +goose Up
CREATE TABLE IF NOT EXISTS urls (
    id SERIAL PRIMARY KEY,
    short_url VARCHAR(8) UNIQUE NOT NULL,
    original_url TEXT UNIQUE NOT NULL
);
-- +goose StatementBegin
SELECT 'up SQL query';
-- +goose StatementEnd

-- +goose Down
DROP TABLE IF EXISTS urls;
-- +goose StatementBegin
SELECT 'down SQL query';
-- +goose StatementEnd
