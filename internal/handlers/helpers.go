package handlers

import (
	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/mastastny/slavoj-web-2025/internal/repository"
)

type Server struct {
	Repo repository.EventRepository
}

func NewServer(repo repository.EventRepository) *Server {
	return &Server{Repo: repo}
}

func renderHTML(c echo.Context, comp templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	return comp.Render(c.Request().Context(), c.Response().Writer)
}
