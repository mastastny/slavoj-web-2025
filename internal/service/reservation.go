package service

import (
	"fmt"
	"log/slog"
	"net/mail"
	"strings"
	"time"

	"github.com/mastastny/slavoj-web-2025/internal/apperrors"
	"github.com/mastastny/slavoj-web-2025/internal/models/reservation"
	"github.com/mastastny/slavoj-web-2025/internal/repository"
	"github.com/mastastny/slavoj-web-2025/internal/service/interface"
)

type Reservation struct {
	eventRepository repository.EventRepository
	emailService    _interface.Email
	locker          _interface.Locker
}

func ConstructReservation(eventRepository repository.EventRepository, email _interface.Email, locker _interface.Locker) *Reservation {
	return &Reservation{
		eventRepository: eventRepository,
		emailService:    email,
		locker:          locker,
	}
}

func (ns *Reservation) Create(newReservation reservation.Service) error {
	if err := validateFields(newReservation); err != nil {
		return err
	}

	start, err := time.Parse(time.RFC3339Nano, newReservation.Start)
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

	area, err := ns.eventRepository.GetCourtByID(newReservation.Area)
	if err != nil {
		return fmt.Errorf("service.Reservation-Create - unable to get area name %w", err)
	}

	lockCode := ns.locker.GetLockerCode(start)

	// todo
	if err := ns.emailService.SendConfirmation(newReservation, area.Name, lockCode, "link"); err != nil {
		return fmt.Errorf("service.Reservation-Create - unable to send confirmation: %w", err)
	}

	return nil
}

func validateFields(r reservation.Service) error {
	if strings.TrimSpace(r.Name) == "" {
		return &apperrors.ValidationError{Message: "Jméno je povinné."}
	}
	if strings.TrimSpace(r.Email) == "" {
		return &apperrors.ValidationError{Message: "E-mail je povinný."}
	}
	if _, err := mail.ParseAddress(r.Email); err != nil {
		return &apperrors.ValidationError{Message: "Zadejte platný e-mail."}
	}
	if r.Area <= 0 {
		return &apperrors.ValidationError{Message: "Neplatné hřiště."}
	}
	if r.PlayerCount <= 0 {
		return &apperrors.ValidationError{Message: "Počet hráčů musí být alespoň 1."}
	}
	return nil
}
