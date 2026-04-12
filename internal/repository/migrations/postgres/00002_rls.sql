-- +goose Up
ALTER TABLE public.courts ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.reservations ENABLE ROW LEVEL SECURITY;
ALTER TABLE public.goose_db_version ENABLE ROW LEVEL SECURITY;
REVOKE ALL ON TABLE public.goose_db_version FROM anon, authenticated;

-- +goose Down
GRANT ALL ON TABLE public.goose_db_version TO anon, authenticated;
ALTER TABLE public.goose_db_version DISABLE ROW LEVEL SECURITY;
ALTER TABLE public.reservations DISABLE ROW LEVEL SECURITY;
ALTER TABLE public.courts DISABLE ROW LEVEL SECURITY;