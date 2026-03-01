package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/mastastny/slavoj-web-2025/internal/views"
)

func GetHome(c echo.Context) error {
	return renderHTML(c, views.Layout("TJ Slavoj Loštice"))
}

func GetAbout(c echo.Context) error {
	return renderHTML(c, views.About())
}

func GetAreals(c echo.Context) error {
	return renderHTML(c, views.Areals())
}

func GetMembership(c echo.Context) error {
	return renderHTML(c, views.Membership())
}

func GetContacts(c echo.Context) error {
	return renderHTML(c, views.Contacts())
}

func GetReservation(c echo.Context) error {
	return renderHTML(c, views.Reservation())
}

func GetModal(c echo.Context) error {
	return renderHTML(c, views.Modal())
}

func GetDocuments(c echo.Context) error {
	return renderHTML(c, views.Documents())
}

func GetHomeContent(c echo.Context) error {
	return renderHTML(c, views.Home())
}
