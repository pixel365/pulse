-- +goose Up
CREATE SCHEMA IF NOT EXISTS pulse;

-- +goose Down
DROP SCHEMA IF EXISTS pulse;
