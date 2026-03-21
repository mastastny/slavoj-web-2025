package handlers

import (
	"errors"

	"github.com/labstack/echo/v4"
	"github.com/mastastny/slavoj-web-2025/internal/apperrors"
	"github.com/mastastny/slavoj-web-2025/internal/models/reservation"
	"github.com/mastastny/slavoj-web-2025/internal/service"
	"github.com/mastastny/slavoj-web-2025/internal/views"
)

type Reservation struct {
	reservationService *service.Reservation
}

func Construct(reservationService *service.Reservation) *Reservation {
	return &Reservation{
		reservationService: reservationService,
	}
}

func (rh *Reservation) GetReservation(c echo.Context) error {
	return renderHTML(c, views.Reservation())
}

func (rh *Reservation) PostReservation(c echo.Context) error {
	var form reservation.Handler

	// missing some form data, this shouldn't be allowed by frontend.
	// todo udelat nejaky chybovy widget
	if err := c.Bind(&form); err != nil {
		return renderHTML(c, views.ReservationError(err.Error()))
	}

	newReservation := handlerToService(form)
	if err := rh.reservationService.Create(newReservation); err != nil {
		if errors.Is(err, apperrors.ErrEventOverlap) {
			return renderHTML(c, views.ReservationError("Nelze vytvořit rezervaci. Ve vybraném časovém rozmezí se již nachází jiná rezervace."))
		}
		if errors.Is(err, apperrors.ErrEventInThePast) {
			return renderHTML(c, views.ReservationError("Rezervaci nelze vytvořit. Zvolený termín musí být v budoucnosti"))
		}
		return renderHTML(c, views.ReservationError("Bohužel právě nejsme schopni vytvořit rezervaci kvůli interní chybě. Zkuste prosím akci opakovat později, nebo kontaktujte administrátora."))
	}

	return renderHTML(c, views.ReservationResult(form))
}

func handlerToService(f reservation.Handler) reservation.Service {
	return reservation.Service{
		Start:       f.Start,
		End:         f.End,
		Area:        f.Area,
		Name:        f.Name,
		Email:       f.Email,
		Phone:       f.Phone,
		PlayerCount: f.PlayerCount,
		Notes:       f.Notes,
		Reminder:    bool(f.Reminder),
	}
}
