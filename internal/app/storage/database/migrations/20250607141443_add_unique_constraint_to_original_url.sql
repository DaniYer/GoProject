-- +goose Up
ALTER TABLE urls ADD CONSTRAINT unique_original_url UNIQUE(original_url);

-- +goose Down
ALTER TABLE urls DROP CONSTRAINT IF EXISTS unique_original_url;
