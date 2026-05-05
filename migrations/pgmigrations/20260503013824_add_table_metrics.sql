-- +goose Up

CREATE TYPE metric_type AS ENUM('gauge', 'counter');

CREATE TABLE IF NOT EXISTS metrics (
    id          SERIAL          PRIMARY KEY,
    name TEXT            NOT NULL UNIQUE,
	type        metric_type     NOT NULL,
	delta       BIGINT,
	value       DOUBLE PRECISION,
    created_at  TIMESTAMPTZ     DEFAULT CURRENT_TIMESTAMP,
    updated_at  TIMESTAMPTZ     DEFAULT CURRENT_TIMESTAMP
);

ALTER TABLE metrics 
ADD CONSTRAINT value_delta_exist CHECK (
    (type = 'gauge' and delta IS NULL and value IS NOT NULL) or 
    (type = 'counter' and delta IS NOT NULL and value IS NULL));

CREATE INDEX IF NOT EXISTS idx_metrics_name ON metrics(name);

-- +goose Down
DROP INDEX IF EXISTS idx_metrics_name;
DROP TABLE IF EXISTS metrics;
DROP TYPE IF EXISTS metric_type;
