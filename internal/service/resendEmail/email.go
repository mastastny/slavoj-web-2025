package resendEmail

import (
	"bytes"
	"context"
	"fmt"
	"log/slog"

	"github.com/mastastny/slavoj-web-2025/internal/config"
	"github.com/mastastny/slavoj-web-2025/internal/models/reservation"
	emailviews "github.com/mastastny/slavoj-web-2025/internal/views/email"
	"github.com/resend/resend-go/v3"
)

type Email struct {
	client *resend.Client
}

func NewEmail(conf config.Config) *Email {
	fmt.Println("XXXX", conf.Resend.APIKey)
	client := resend.NewClient(conf.Resend.APIKey)
	return &Email{client: client}
}

func (e *Email) SendConfirmation(r reservation.Service, courtName string, lockCode string, cancelLink string) error {
	var buf bytes.Buffer
	if err := emailviews.ReservationConfirmation(r, courtName, lockCode, cancelLink).Render(context.Background(), &buf); err != nil {
		return fmt.Errorf("resendEmail.SendConfirmation: render template: %w", err)
	}

	params := &resend.SendEmailRequest{
		From:    "rezervace@slavojlostice.cz",
		To:      []string{r.Email},
		Subject: "Potvrzení rezervace — TJ Slavoj Loštice",
		Html:    buf.String(),
	}

	sent, err := e.client.Emails.Send(params)
	if err != nil {
		return fmt.Errorf("resendEmail.SendConfirmation: send: %w", err)
	}

	slog.Info("confirmation email sent", "id", sent.Id, "to", r.Email)
	return nil
}

func (e *Email) RegisterRemainder(reservation reservation.Service) error {
	slog.Warn("resend email service: using function RegisterRemainder which is not yet implemented. Reservation: %#v", reservation)
	return nil
}
