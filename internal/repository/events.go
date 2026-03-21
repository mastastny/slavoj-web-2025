package repository

import (
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"time"

	"github.com/mastastny/slavoj-web-2025/internal/apperrors"
	"github.com/mastastny/slavoj-web-2025/internal/models"
	"github.com/mastastny/slavoj-web-2025/internal/models/reservation"
	"github.com/mattn/go-sqlite3"
)

//go:embed queries/get_events_by_court_and_range.sql
var getEventsByCourtAndRange string

//go:embed queries/create_event.sql
var createEvent string

type EventRepository interface {
	GetEventsByCourtAndRange(courtID, startStr, endStr string) ([]models.Event, error)
	CreateEvent(event reservation.Service) error
}

type SqliteEventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) EventRepository {
	return &SqliteEventRepository{db: db}
}

func (r *SqliteEventRepository) CreateEvent(event reservation.Service) error {
	_, err := r.db.Exec(createEvent, event.Area, event.Start, event.End, event.Name, event.Email)
	if err != nil {
		var sqliteErr sqlite3.Error
		if errors.As(err, &sqliteErr) && sqliteErr.ExtendedCode == sqlite3.ErrConstraintTrigger {
			return apperrors.ErrEventOverlap
		}
		return fmt.Errorf("EventRepository-CreateEvent: %w", err)
	}
	return nil
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
