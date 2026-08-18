-- +goose Up
-- Single-row table: id is forced to TRUE so INSERT ... ON CONFLICT (id) DO UPDATE
-- always upserts the one current locker OAuth token, shared across all Lambda
-- instances (unlike env vars or in-process memory, which don't survive a cold
-- start or a swap to a different concurrent execution environment).
CREATE TABLE IF NOT EXISTS locker_token (
    id            BOOLEAN PRIMARY KEY DEFAULT TRUE CHECK (id),
    access_token  TEXT NOT NULL,
    refresh_token TEXT NOT NULL,
    lifespan_secs INTEGER NOT NULL,
    issued_at     TIMESTAMPTZ NOT NULL
);

ALTER TABLE public.locker_token ENABLE ROW LEVEL SECURITY;
REVOKE ALL ON TABLE public.locker_token FROM anon, authenticated;

-- +goose Down
DROP TABLE IF EXISTS locker_token;
