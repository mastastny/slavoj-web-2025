package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

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
	pool *pgxpool.Pool
}

// NewSupabaseLockerTokenRepository assumes the locker_token table has
// already been migrated (NewSupabaseEventRepository runs the migrations for
// the whole "migrations/postgres" directory, which includes it).
func NewSupabaseLockerTokenRepository(databaseURL string) (*SupabaseLockerTokenRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("NewSupabaseLockerTokenRepository: connect: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("NewSupabaseLockerTokenRepository: ping: %w", err)
	}
	return &SupabaseLockerTokenRepository{pool: pool}, nil
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
