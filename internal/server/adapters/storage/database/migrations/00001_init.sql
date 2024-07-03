-- +goose Up
CREATE TABLE IF NOT EXISTS metrics
(
    id            INT PRIMARY KEY GENERATED ALWAYS AS IDENTITY,
    name          varchar(255) not null,
    type          varchar(255) not null,
    delta         bigint,
    value         double precision,
    created_at    timestamp without time zone NOT NULL DEFAULT (current_timestamp AT TIME ZONE 'UTC')
    CHECK (delta IS NOT NULL OR value IS NOT NULL)
    );

CREATE INDEX IF NOT EXISTS name_idx ON metrics (name);
CREATE INDEX IF NOT EXISTS type_idx ON metrics (type);

-- +goose Down
DROP TABLE metrics;