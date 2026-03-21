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

CREATE TRIGGER IF NOT EXISTS prevent_overlap_insert
BEFORE INSERT ON reservations
FOR EACH ROW
BEGIN
    SELECT RAISE(ABORT, 'Reservation overlaps with existing one')
    WHERE EXISTS (
        SELECT 1 FROM reservations
        WHERE court_id = NEW.court_id
          AND start_at < NEW.end_at
          AND end_at   > NEW.start_at
    );
END;

CREATE TRIGGER IF NOT EXISTS prevent_overlap_update
BEFORE UPDATE ON reservations
FOR EACH ROW
BEGIN
    SELECT RAISE(ABORT, 'Reservation overlaps with existing one')
    WHERE EXISTS (
        SELECT 1 FROM reservations
        WHERE court_id = NEW.court_id
          AND id      != NEW.id
          AND start_at < NEW.end_at
          AND end_at   > NEW.start_at
    );
END;
