-- +goose Up
CREATE TABLE IF NOT EXISTS courts (
    id   INTEGER PRIMARY KEY,
    name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS reservations (
    id       BIGSERIAL PRIMARY KEY,
    court_id INTEGER      NOT NULL REFERENCES courts(id),
    start_at TIMESTAMPTZ  NOT NULL,
    end_at   TIMESTAMPTZ  NOT NULL,
    name     TEXT         NOT NULL,
    email    TEXT,
    CONSTRAINT uq_slot UNIQUE (court_id, start_at, end_at)
);

CREATE INDEX IF NOT EXISTS idx_res_by_court_time
    ON reservations(court_id, start_at, end_at);

-- +goose Down
DROP INDEX IF EXISTS idx_res_by_court_time;
DROP TABLE IF EXISTS reservations;
DROP TABLE IF EXISTS courts;
