package handlers

import (
	"github.com/labstack/echo/v4"
	"github.com/mastastny/slavoj-web-2025/internal/models/reservation"
	"github.com/mastastny/slavoj-web-2025/internal/views"
	emailviews "github.com/mastastny/slavoj-web-2025/internal/views/email"
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

var novinka1 = views.NovinkaDef{
	Title: "Úspěšný sportovec okresu Šumperk za rok 2025",
	Text: `Za rok 2025 byli vyhodnoceni naši sportovci TJ Slavoj Loštice z.s. v rámci výzvy ` +
		`"Úspěšný sportovec okresu Šumperk za rok 2025". Oceněno bylo v kategorii dospělých ` +
		`družstev družstvo volejbalu žen, které se v loňském roce probojovalo do 2. ligy, ` +
		`a v kategorii osobnost v oblasti sportu Vladimír Kindl, dlouholetý předseda TJ Slavoj. ` +
		`Ten obdržel i ocenění Výboru ČUS.`,
	Photos: []views.NovinkaPhoto{
		{Src: "/images/novinka1/selfie.jpg", Alt: "Ocenění sportovci"},
		{Src: "/images/novinka1/vsichni.jpg", Alt: "Společná fotka"},
		{Src: "/images/novinka1/kindl-uznani.jpg", Alt: "Čestné uznání Vladimíru Kindlovi"},
		{Src: "/images/novinka1/kindl-diplom.jpg", Alt: "Diplom Vladimíra Kindla"},
	},
}

func GetNovinka1(c echo.Context) error {
	if c.Request().Header.Get("HX-Request") == "true" {
		return renderHTML(c, views.Novinka(novinka1))
	}
	return renderHTML(c, views.Layout(novinka1.Title+" | TJ Slavoj Loštice", views.Novinka(novinka1)))
}

func GetEmailPreview(c echo.Context) error {
	mock := reservation.Service{
		Name:        "Jan Novák",
		Email:       "jan.novak@example.com",
		Start:       "2026-05-10T10:00:00Z",
		End:         "2026-05-10T12:00:00Z",
		PlayerCount: 4,
		Notes:       "Prosím o zapůjčení raket.",
	}
	return renderHTML(c, emailviews.ReservationConfirmation(mock, "tenisový kurt č. 1 (antuka)", "4729", "https://slavojlostice.cz/reservation/cancel/abc123"))
}

func GetHomeContent(c echo.Context) error {
	if c.Request().Header.Get("HX-Request") == "true" {
		return renderHTML(c, views.Home())
	}
	return renderHTML(c, views.Layout("TJ Slavoj Loštice", views.Home()))
}
