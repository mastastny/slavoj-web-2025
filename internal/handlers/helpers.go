package handlers

import (
	"database/sql"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/mastastny/slavoj-web-2025/internal/repository"
)

type Server struct {
	DB   *sql.DB
	Repo repository.EventRepository
}

func NewServer(db *sql.DB) *Server {
	return &Server{
		DB:   db,
		Repo: repository.NewEventRepository(db),
	}
}

func renderHTML(c echo.Context, comp templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	return comp.Render(c.Request().Context(), c.Response().Writer)
}
