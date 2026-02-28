PRAGMA foreign_keys = ON;

CREATE TABLE IF NOT EXISTS courts (
                                      id   INTEGER PRIMARY KEY,
                                      name TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS reservations (
                                            id       INTEGER PRIMARY KEY,
                                            court_id INTEGER NOT NULL,
                                            start_at TEXT    NOT NULL, -- ISO8601 v UTC (RFC3339)
                                            end_at   TEXT    NOT NULL,
                                            name     TEXT    NOT NULL,
                                            email    TEXT,
                                            CONSTRAINT uq_slot UNIQUE (court_id, start_at, end_at),
    FOREIGN KEY(court_id) REFERENCES courts(id)
    );

CREATE INDEX IF NOT EXISTS idx_res_by_court_time
    ON reservations(court_id, start_at, end_at);
