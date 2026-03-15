package handlers

import (
	"fmt"
	"strconv"

	"github.com/labstack/echo/v4"
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
	if err := c.Bind(&form); err != nil {
		return renderHTML(c, views.ReservationResult(err.Error()))
	}

	newReservation := handlerToService(form)
	rh.reservationService.Create(newReservation)

	fmt.Println(form)
	fmt.Print("dobry den")
	if c.FormValue("name") == "Marek" {
		return renderHTML(c, views.ReservationResult("Vitejte pane"+
			"\n Jmeno: "+form.Name+
			"\n Hriste: "+fmt.Sprintf("%d", form.Area)+
			"\n reminder: "+strconv.FormatBool(bool(form.Reminder))+
			"\n start: "+form.Start))
	}
	return renderHTML(c, views.ReservationResult("baad"))
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
