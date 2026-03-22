package service

import "github.com/mastastny/slavoj-web-2025/internal/models/reservation"

type Email interface {
	SendConfirmation(r reservation.Service, courtName string, lockCode string, cancelLink string) error
	RegisterRemainder(reservation reservation.Service) error
}
