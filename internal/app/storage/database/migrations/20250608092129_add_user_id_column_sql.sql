-- +goose Up
ALTER TABLE urls ADD COLUMN user_id VARCHAR(36);

-- +goose Down
ALTER TABLE urls DROP COLUMN IF EXISTS user_id;