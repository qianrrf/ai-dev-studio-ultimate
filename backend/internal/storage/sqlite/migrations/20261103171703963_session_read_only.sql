-- +goose Up
ALTER TABLE sessions ADD COLUMN read_only INTEGER NOT NULL DEFAULT 0 CHECK (read_only IN (0, 1));

-- +goose Down
ALTER TABLE sessions DROP COLUMN read_only;
