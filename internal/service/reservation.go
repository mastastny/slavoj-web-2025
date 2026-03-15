package service

import (
	"fmt"

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

func (ns *Reservation) Create(newReservation reservation.Service) {
	fmt.Println("Service, vytvarim rezervaci")
	fmt.Printf("%#v\n", newReservation)
}
