package repository

import (
	"github.com/mastastny/slavoj-web-2025/internal/models"
	"github.com/mastastny/slavoj-web-2025/internal/models/reservation"
)

type EventRepository interface {
	GetEventsByCourtAndRange(courtID, startStr, endStr string) ([]models.Event, error)
	CreateEvent(event reservation.Service) (int64, error)
	DeleteEvent(id int64) error
	GetCourtByID(id int) (models.Court, error)
	GetCourts() []models.Court
}
