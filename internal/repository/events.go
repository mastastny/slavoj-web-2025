package repository

import (
	"database/sql"
	"time"

	"github.com/mastastny/slavoj-web-2025/internal/models"
)

type EventRepository struct {
	DB *sql.DB
}

func NewEventRepository(db *sql.DB) *EventRepository {
	return &EventRepository{DB: db}
}

func (r *EventRepository) GetEventsByCourtAndRange(courtID, startStr, endStr string) ([]models.Event, error) {
	rows, err := r.DB.Query(`
		SELECT name, start_at, end_at
		FROM reservations
		WHERE court_id = ?
		  AND start_at >= ?
		  AND end_at   <= ?
		ORDER BY start_at
	`, courtID, startStr, endStr)
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
