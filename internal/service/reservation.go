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
	linkCoder       *LinkCoder
	publicDomain    string
}

func ConstructReservation(eventRepository repository.EventRepository, email _interface.Email, locker _interface.Locker, linkCoder *LinkCoder, publicDomain string) *Reservation {
	return &Reservation{
		eventRepository: eventRepository,
		emailService:    email,
		locker:          locker,
		linkCoder:       linkCoder,
		publicDomain:    publicDomain,
	}
}

func (ns *Reservation) Cancel(encodedID string) error {
	id, err := ns.linkCoder.Decode(encodedID)
	if err != nil {
		return fmt.Errorf("Reservation.Cancel - invalid id: %w", err)
	}
	if err := ns.eventRepository.DeleteEvent(id); err != nil {
		return fmt.Errorf("Reservation.Cancel - unable to delete event: %w", err)
	}
	slog.Info("Reservation cancelled", "id", id)
	return nil
}

func (ns *Reservation) Create(newReservation reservation.Service) error {
	if err := validateFields(newReservation); err != nil {
		return err
	}

	tz, err := time.LoadLocation("Europe/Prague")
	if err != nil {
		return fmt.Errorf("service.Reservation-Create - cannot load timezone: %w", err)
	}

	parsedStart, err := time.Parse(time.RFC3339Nano, newReservation.Start)
	if err != nil {
		return fmt.Errorf("service.Reservation-Create - invalid start time: %w", err)
	}
	// FullCalendar sends fake-UTC times where UTC components = Prague local time
	start := time.Date(parsedStart.Year(), parsedStart.Month(), parsedStart.Day(),
		parsedStart.Hour(), parsedStart.Minute(), parsedStart.Second(), parsedStart.Nanosecond(), tz)

	if start.Before(time.Now()) {
		return apperrors.ErrEventInThePast
	}

	parsedEnd, err := time.Parse(time.RFC3339Nano, newReservation.End)
	if err != nil {
		return fmt.Errorf("service.Reservation-Create - invalid end time: %w", err)
	}
	end := time.Date(parsedEnd.Year(), parsedEnd.Month(), parsedEnd.Day(),
		parsedEnd.Hour(), parsedEnd.Minute(), parsedEnd.Second(), parsedEnd.Nanosecond(), tz)

	eventId, err := ns.eventRepository.CreateEvent(newReservation)
	if err != nil {
		return fmt.Errorf("service.Reservation-Create - unable to create an event: %w", err)
	}

	slog.Info("Created new event", "reservation", fmt.Sprintf("%#v", newReservation))

	area, err := ns.eventRepository.GetCourtByID(newReservation.Area)
	if err != nil {
		return fmt.Errorf("service.Reservation-Create - unable to get area name %w", err)
	}

	//lockerCode := ns.locker.GetLockerCode(start)
	lockerCode := ns.locker.CreatePasscode(start, end)

	cancelLink := ns.publicDomain + "/reservation/cancel/" + ns.linkCoder.Encode(eventId)

	if err := ns.emailService.SendConfirmation(newReservation, area.Name, lockerCode, cancelLink); err != nil {
		return fmt.Errorf("service.Reservation-Create - unable to sendPasscodeToLock confirmation: %w", err)
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
