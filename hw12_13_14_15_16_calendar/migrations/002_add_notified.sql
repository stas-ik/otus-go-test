-- +goose Up
ALTER TABLE events ADD COLUMN notified BOOLEAN DEFAULT FALSE;

-- +goose Down
ALTER TABLE events DROP COLUMN notified;
