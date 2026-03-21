package service

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/mastastny/slavoj-web-2025/internal/apperrors"
	"github.com/mastastny/slavoj-web-2025/internal/models/reservation"
	"github.com/mastastny/slavoj-web-2025/internal/repository"
)

type Reservation struct {
	eventRepository repository.EventRepository
}

func ConstructReservation(eventRepository repository.EventRepository) *Reservation {
	return &Reservation{
		eventRepository: eventRepository,
	}
}

func (ns *Reservation) Create(newReservation reservation.Service) error {
	start, err := time.Parse(time.RFC3339, newReservation.Start)
	if err != nil {
		return fmt.Errorf("service.Reservation-Create - invalid start time: %w", err)
	}
	if start.Before(time.Now()) {
		return apperrors.ErrEventInThePast
	}

	fmt.Println("Service, vytvarim rezervaci")

	if err := ns.eventRepository.CreateEvent(newReservation); err != nil {
		return fmt.Errorf("service.Reservation-Create - unable to create an event: %w", err)
	}

	slog.Info("Created new event", "reservation", fmt.Sprintf("%#v", newReservation))
	return nil
}

func validateFields(reservation reservation.Service) {
	// todo
}
