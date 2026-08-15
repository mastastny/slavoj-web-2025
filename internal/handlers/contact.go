package handlers

import (
	"log/slog"

	"github.com/labstack/echo/v4"
	_interface "github.com/mastastny/slavoj-web-2025/internal/service/interface"
	"github.com/mastastny/slavoj-web-2025/internal/views"
)

type Contact struct {
	emailService _interface.Email
}

func ConstructContact(emailService _interface.Email) *Contact {
	return &Contact{emailService: emailService}
}

type contactFormData struct {
	Name    string `form:"name" binding:"required"`
	Email   string `form:"email" binding:"required"`
	Subject string `form:"subject"`
	Message string `form:"message" binding:"required"`
}

func (ch *Contact) PostContact(c echo.Context) error {
	var form contactFormData
	if err := c.Bind(&form); err != nil {
		return renderHTML(c, views.ContactError("Nepodařilo se zpracovat formulář."))
	}

	if form.Name == "" || form.Email == "" || form.Message == "" {
		return renderHTML(c, views.ContactError("Vyplňte prosím všechna pole."))
	}

	if err := ch.emailService.SendContactMessage(form.Name, form.Email, form.Subject, form.Message); err != nil {
		slog.Error("PostContact: send failed", "err", err)
		return renderHTML(c, views.ContactError("Zprávu se nepodařilo odeslat. Zkuste to prosím později."))
	}

	return renderHTML(c, views.ContactSuccess())
}
