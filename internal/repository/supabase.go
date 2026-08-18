package repository

import (
	"context"
	"database/sql"
	"embed"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/mastastny/slavoj-web-2025/internal/apperrors"
	"github.com/mastastny/slavoj-web-2025/internal/models"
	"github.com/mastastny/slavoj-web-2025/internal/models/reservation"
	"github.com/pressly/goose/v3"
)

//go:embed queries/supabase/get_events_by_court_and_range.sql
var supabaseGetEventsByCourtAndRange string

//go:embed queries/supabase/check_overlap.sql
var supabaseCheckOverlap string

//go:embed queries/supabase/create_event.sql
var supabaseCreateEvent string

//go:embed queries/supabase/delete_event.sql
var supabaseDeleteEvent string

//go:embed migrations/postgres/*.sql
var postgresMigrations embed.FS

type SupabaseEventRepository struct {
	pool   *pgxpool.Pool
	courts []models.Court
}

func NewSupabaseEventRepository(databaseURL string, courts []models.Court) (*SupabaseEventRepository, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := sql.Open("pgx", databaseURL)
	if err != nil {
		return nil, fmt.Errorf("NewSupabaseEventRepository: open for migrations: %w", err)
	}
	defer db.Close()

	goose.SetBaseFS(postgresMigrations)
	if err := goose.SetDialect("postgres"); err != nil {
		return nil, fmt.Errorf("NewSupabaseEventRepository: set dialect: %w", err)
	}
	if err := goose.Up(db, "migrations/postgres"); err != nil {
		return nil, fmt.Errorf("NewSupabaseEventRepository: migrate: %w", err)
	}

	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return nil, fmt.Errorf("NewSupabaseEventRepository: connect: %w", err)
	}
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("NewSupabaseEventRepository: ping: %w", err)
	}

	for _, c := range courts {
		_, err := pool.Exec(context.Background(), `INSERT INTO courts (id, name) VALUES ($1, $2) ON CONFLICT (id) DO NOTHING`, c.ID, c.Name)
		if err != nil {
			return nil, fmt.Errorf("NewSupabaseEventRepository: seed court %d: %w", c.ID, err)
		}
	}
	return &SupabaseEventRepository{pool: pool, courts: courts}, nil
}

// Pool exposes the connection pool so other repositories against the same
// database (e.g. SupabaseLockerTokenRepository) can share it instead of
// opening their own - Supabase poolers cap total concurrent clients.
func (r *SupabaseEventRepository) Pool() *pgxpool.Pool {
	return r.pool
}

func (r *SupabaseEventRepository) GetCourts() []models.Court {
	return r.courts
}

func (r *SupabaseEventRepository) GetCourtByID(id int) (models.Court, error) {
	for _, c := range r.courts {
		if c.ID == id {
			return c, nil
		}
	}
	return models.Court{}, fmt.Errorf("court with id %d not found", id)
}

func (r *SupabaseEventRepository) GetEventsByCourtAndRange(courtID, startStr, endStr string) ([]models.Event, error) {
	ctx := context.Background()
	rows, err := r.pool.Query(ctx, supabaseGetEventsByCourtAndRange, courtID, startStr, endStr)
	if err != nil {
		return nil, fmt.Errorf("SupabaseEventRepository-GetEventsByCourtAndRange: %w", err)
	}
	defer rows.Close()

	out := make([]models.Event, 0)
	for rows.Next() {
		var title string
		var start, end time.Time
		if err := rows.Scan(&title, &start, &end); err != nil {
			return nil, err
		}
		out = append(out, models.Event{Title: title, Start: start, End: end})
	}
	return out, nil
}

func (r *SupabaseEventRepository) CreateEvent(event reservation.Service) (int64, error) {
	ctx := context.Background()
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return -1, fmt.Errorf("SupabaseEventRepository-CreateEvent: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	var count int
	if err := tx.QueryRow(ctx, supabaseCheckOverlap, event.Area, event.End, event.Start).Scan(&count); err != nil {
		return -1, fmt.Errorf("SupabaseEventRepository-CreateEvent: check overlap: %w", err)
	}
	if count > 0 {
		return -1, apperrors.ErrEventOverlap
	}

	var id int64
	if err := tx.QueryRow(ctx, supabaseCreateEvent, event.Area, event.Start, event.End, event.Name, event.Email).Scan(&id); err != nil {
		return -1, fmt.Errorf("SupabaseEventRepository-CreateEvent: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return -1, fmt.Errorf("SupabaseEventRepository-CreateEvent: commit: %w", err)
	}
	return id, nil
}

func (r *SupabaseEventRepository) DeleteEvent(id int64) error {
	ctx := context.Background()
	if _, err := r.pool.Exec(ctx, supabaseDeleteEvent, id); err != nil {
		return fmt.Errorf("SupabaseEventRepository-DeleteEvent: %w", err)
	}
	return nil
}
