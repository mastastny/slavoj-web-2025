package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/mastastny/slavoj-web-2025/internal/views"
)

func GetHome(c echo.Context) error {
	return renderHTML(c, views.Layout("TJ Slavoj Loštice", views.Home()))
}

func GetAbout(c echo.Context) error {
	if c.Request().Header.Get("HX-Request") == "true" {
		return renderHTML(c, views.About())
	}
	return renderHTML(c, views.Layout("TJ Slavoj Loštice", views.About()))
}

func GetAreals(c echo.Context) error {
	if c.Request().Header.Get("HX-Request") == "true" {
		return renderHTML(c, views.Areals())
	}
	return renderHTML(c, views.Layout("TJ Slavoj Loštice", views.Areals()))
}

func GetMembership(c echo.Context) error {
	if c.Request().Header.Get("HX-Request") == "true" {
		return renderHTML(c, views.Membership())
	}
	return renderHTML(c, views.Layout("Členství | TJ Slavoj Loštice", views.Membership()))
}

func GetContacts(c echo.Context) error {
	if c.Request().Header.Get("HX-Request") == "true" {
		return renderHTML(c, views.Contacts())
	}
	return renderHTML(c, views.Layout("Kontakty | TJ Slavoj Loštice", views.Contacts()))
}

func GetModal(c echo.Context) error {
	return renderHTML(c, views.Modal())
}

func GetModalBookingForm(c echo.Context) error {
	return renderHTML(c, views.ModalBookingForm())
}

func GetDocuments(c echo.Context) error {
	if c.Request().Header.Get("HX-Request") == "true" {
		return renderHTML(c, views.Documents())
	}
	return renderHTML(c, views.Layout("Dokumenty | TJ Slavoj Loštice", views.Documents()))
}

func GetHomeContent(c echo.Context) error {
	return renderHTML(c, views.Home())
}
