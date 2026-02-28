package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func (s *Server) GetEvents(c echo.Context) error {
	courtID := c.QueryParam("court_id")
	if courtID == "" {
		courtID = "1"
	}
	startStr := c.QueryParam("start")
	endStr := c.QueryParam("end")

	events, err := s.Repo.GetEventsByCourtAndRange(courtID, startStr, endStr)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, events)
}
