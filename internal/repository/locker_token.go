package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// pgxQuerier is the subset of *pgxpool.Pool this repository needs, so it can
// share a pool with other repositories instead of opening its own (Supabase
// poolers cap total concurrent clients, so extra pools are expensive).
type pgxQuerier interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
}

// LockerToken is the Sciener/TTLock OAuth access+refresh token pair the
// locker service uses to authenticate against the lock API.
type LockerToken struct {
	AccessToken   string
	RefreshToken  string
	TokenLifespan int // seconds
	IssuedAt      time.Time
}

// LockerTokenRepository persists the current locker token so it survives
// Lambda cold starts and is shared across concurrent execution environments.
// Without this, each cold start would fall back to the refresh_token baked
// into the deploy-time env var, which TTLock invalidates as soon as any
// other instance has already rotated it - producing "invalid grant" errors.
type LockerTokenRepository interface {
	// LoadLockerToken returns the persisted token, or found=false if none has
	// been stored yet.
	LoadLockerToken() (token LockerToken, found bool, err error)
	SaveLockerToken(token LockerToken) error
}

type SupabaseLockerTokenRepository struct {
	pool pgxQuerier
}

// NewSupabaseLockerTokenRepository takes an already-connected pool (share the
// one NewSupabaseEventRepository opens) rather than opening its own -
// Supabase poolers cap total concurrent clients (e.g. the session pooler at
// 15), so a second pool per Lambda container can exhaust that budget.
// It assumes the locker_token table has already been migrated, which
// NewSupabaseEventRepository does for the whole "migrations/postgres"
// directory.
func NewSupabaseLockerTokenRepository(pool *pgxpool.Pool) *SupabaseLockerTokenRepository {
	return &SupabaseLockerTokenRepository{pool: pool}
}

func (r *SupabaseLockerTokenRepository) LoadLockerToken() (LockerToken, bool, error) {
	ctx := context.Background()
	var t LockerToken
	err := r.pool.QueryRow(ctx,
		`SELECT access_token, refresh_token, lifespan_secs, issued_at FROM locker_token WHERE id = TRUE`,
	).Scan(&t.AccessToken, &t.RefreshToken, &t.TokenLifespan, &t.IssuedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return LockerToken{}, false, nil
		}
		return LockerToken{}, false, fmt.Errorf("LoadLockerToken: %w", err)
	}
	return t, true, nil
}

func (r *SupabaseLockerTokenRepository) SaveLockerToken(token LockerToken) error {
	ctx := context.Background()
	_, err := r.pool.Exec(ctx, `
		INSERT INTO locker_token (id, access_token, refresh_token, lifespan_secs, issued_at)
		VALUES (TRUE, $1, $2, $3, $4)
		ON CONFLICT (id) DO UPDATE SET
			access_token  = EXCLUDED.access_token,
			refresh_token = EXCLUDED.refresh_token,
			lifespan_secs = EXCLUDED.lifespan_secs,
			issued_at     = EXCLUDED.issued_at
	`, token.AccessToken, token.RefreshToken, token.TokenLifespan, token.IssuedAt)
	if err != nil {
		return fmt.Errorf("SaveLockerToken: %w", err)
	}
	return nil
}
