package sendGridEmail

import (
	"github.com/mastastny/slavoj-web-2025/internal/config"
	"github.com/sendgrid/sendgrid-go"
)

type Email struct {
	client *sendgrid.Client
	config config.Config
}

func NewEmail(conf config.Config) *Email {
	client := sendgrid.NewSendClient(conf.SendGrid.APIKey)
	return &Email{
		client: client,
		config: conf,
	}
}
