package main

import (
	//"net/http"

	"context"
	"database/sql"
	"net/http"
	"os"
	"time"

	"github.com/a-h/templ"
	"github.com/labstack/echo/v4"
	"github.com/mastastny/slavoj-web-2025/views"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	echoadapter "github.com/awslabs/aws-lambda-go-api-proxy/echo"
)

var echoLambda *echoadapter.EchoLambdaV2

func GetHome(c echo.Context) error {
	//return c.Render(http.StatusOK, "home", nil)
	page := views.Layout("TJ Slavoj Loštice")
	return renderHTML(c, page)
}

func GetAbout(c echo.Context) error {
	//return c.Render(http.StatusOK, "home", nil)
	page := views.About()
	return renderHTML(c, page)
}

func GetAreals(c echo.Context) error {
	//return c.Render(http.StatusOK, "home", nil)
	page := views.Areals()
	return renderHTML(c, page)
}

func GetContacts(c echo.Context) error {
	//return c.Render(http.StatusOK, "home", nil)
	page := views.Contacts()
	return renderHTML(c, page)
}

func GetReservation(c echo.Context) error {
	//return c.Render(http.StatusOK, "home", nil)
	page := views.Reservation()
	return renderHTML(c, page)
}

func GetModal(c echo.Context) error {
	//return c.Render(http.StatusOK, "home", nil)
	page := views.Modal()
	return renderHTML(c, page)
}

// helper to render templ components in Echo
func renderHTML(c echo.Context, comp templ.Component) error {
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMETextHTMLCharsetUTF8)
	return comp.Render(c.Request().Context(), c.Response().Writer)
}

type Server struct {
	DB *sql.DB
}

func NewServer(db *sql.DB) *Server {
	return &Server{DB: db}
}

func (s *Server) GetEvents(c echo.Context) error {
	courtID := c.QueryParam("court_id")
	if courtID == "" {
		courtID = "1"
	}
	startStr := c.QueryParam("start")
	endStr := c.QueryParam("end")
	// FullCalendar posílá ISO8601; u nás v DB je UTC ISO8601 (TEXT)

	rows, err := s.DB.Query(`
			SELECT name, start_at, end_at
			FROM reservations
			WHERE court_id = ?
			  AND start_at >= ?
			  AND end_at   <= ?
			ORDER BY start_at
		`, courtID, startStr, endStr)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}
	defer rows.Close()

	out := make([]Event, 0)
	for rows.Next() {
		var title, s, e2 string
		if err := rows.Scan(&title, &s, &e2); err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		// Parse z ISO8601 do time.Time (UTC)
		st, err := time.Parse(time.RFC3339, s)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		en, err := time.Parse(time.RFC3339, e2)
		if err != nil {
			return c.String(http.StatusInternalServerError, err.Error())
		}
		out = append(out, Event{Title: title, Start: st, End: en})
	}
	return c.JSON(http.StatusOK, out)

	//m := []dtos.Event{{"Mara", time.Now(), time.Now().Add(time.Hour)}}
	//b, _ := json.Marshal(m)
	//
	//return c.JSONBlob(http.StatusOK, b)
}

func main() {

	db := Init()
	defer db.Close()
	server := NewServer(db)

	e := echo.New()
	e.Static("/", "static")
	//index := views.Index()
	e.GET("/", GetHome)
	e.GET("/about", GetAbout)
	e.GET("/areals", GetAreals)
	e.GET("/reservation", GetReservation)
	e.GET("/contacts", GetContacts)
	e.GET("/modal", GetModal)

	e.GET("/api/events", server.GetEvents)

	// 2. Detect if running in AWS Lambda environment
	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		// We are in Lambda, so do not start Echo in the usual way.
		// Initialize the Echo adapter for API Gateway V2 (HTTP API)
		echoLambda = echoadapter.NewV2(e)
		// Start the Lambda event processing loop with our handler function
		lambda.Start(handler)
	} else {
		// Not in Lambda (running locally or in another environment), start Echo normally
		e.Logger.Fatal(e.Start(":8080"))
	}

}

// 3. Lambda handler function for API Gateway HTTP API (v2) events
func handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	// Proxy the incoming API Gateway request to the Echo instance and return the response
	return echoLambda.ProxyWithContext(ctx, req)
}
