package repository

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/mastastny/slavoj-web-2025/internal/apperrors"
	"github.com/mastastny/slavoj-web-2025/internal/models"
	"github.com/mastastny/slavoj-web-2025/internal/models/reservation"
	"github.com/mattn/go-sqlite3"
)

//go:embed init.sql
var sqliteInitSQL string

//go:embed queries/get_events_by_court_and_range.sql
var getEventsByCourtAndRange string

//go:embed queries/create_event.sql
var createEvent string

//go:embed queries/delete_event.sql
var deleteEvent string

type SqliteEventRepository struct {
	db     *sql.DB
	courts []models.Court
}

func NewSqliteEventRepository(courts []models.Court) EventRepository {
	dbPath := os.Getenv("DATABASE_PATH")
	if dbPath == "" {
		dbPath = "club.sqlite"
	}
	db, err := sql.Open("sqlite3", "file:"+dbPath+"?_foreign_keys=on&_busy_timeout=5000")
	if err != nil {
		panic(fmt.Errorf("NewSqliteEventRepository: open: %w", err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := db.ExecContext(ctx, sqliteInitSQL); err != nil {
		panic(fmt.Errorf("NewSqliteEventRepository: migrate: %w", err))
	}

	for _, c := range courts {
		if _, err := db.Exec(`INSERT OR IGNORE INTO courts (id, name) VALUES (?, ?)`, c.ID, c.Name); err != nil {
			panic(fmt.Errorf("NewSqliteEventRepository: seed court %d: %w", c.ID, err))
		}
	}
	return &SqliteEventRepository{db: db, courts: courts}
}

func (r *SqliteEventRepository) DeleteEvent(id int64) error {
	_, err := r.db.Exec(deleteEvent, id)
	if err != nil {
		return fmt.Errorf("SqliteEventRepository-DeleteEvent: %w", err)
	}
	return nil
}

func (r *SqliteEventRepository) GetCourts() []models.Court {
	return r.courts
}

func (r *SqliteEventRepository) GetCourtByID(id int) (models.Court, error) {
	for _, c := range r.courts {
		if c.ID == id {
			return c, nil
		}
	}
	return models.Court{}, fmt.Errorf("court with id %d not found", id)
}

func (r *SqliteEventRepository) CreateEvent(event reservation.Service) (int64, error) {
	result, err := r.db.Exec(createEvent, event.Area, event.Start, event.End, event.Name, event.Email)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintTrigger {
			return -1, apperrors.ErrEventOverlap
		}
		return -1, fmt.Errorf("SqliteEventRepository-CreateEvent: %w", err)
	}

	id, _ := result.LastInsertId()
	return id, nil
}

func (r *SqliteEventRepository) GetEventsByCourtAndRange(courtID, startStr, endStr string) ([]models.Event, error) {
	rows, err := r.db.Query(getEventsByCourtAndRange, courtID, startStr, endStr)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]models.Event, 0)
	for rows.Next() {
		var title, s, e string
		if err := rows.Scan(&title, &s, &e); err != nil {
			return nil, err
		}
		st, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return nil, err
		}
		en, err := time.Parse(time.RFC3339, e)
		if err != nil {
			return nil, err
		}
		out = append(out, models.Event{Title: title, Start: st, End: en})
	}
	return out, nil
}
