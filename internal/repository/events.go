package repository

import (
	"database/sql"
	_ "embed"
	"time"

	"github.com/mastastny/slavoj-web-2025/internal/models"
)

//go:embed queries/get_events_by_court_and_range.sql
var getEventsByCourtAndRange string

type EventRepository interface {
	GetEventsByCourtAndRange(courtID, startStr, endStr string) ([]models.Event, error)
}

type sqliteEventRepository struct {
	db *sql.DB
}

func NewEventRepository(db *sql.DB) EventRepository {
	return &sqliteEventRepository{db: db}
}

func (r *sqliteEventRepository) GetEventsByCourtAndRange(courtID, startStr, endStr string) ([]models.Event, error) {
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
