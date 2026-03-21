package main

import (
	"context"
	"fmt"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	echoadapter "github.com/awslabs/aws-lambda-go-api-proxy/echo"
	"github.com/labstack/echo/v4"
	config "github.com/mastastny/slavoj-web-2025/internal/config"
	"github.com/mastastny/slavoj-web-2025/internal/database"
	"github.com/mastastny/slavoj-web-2025/internal/handlers"
	"github.com/mastastny/slavoj-web-2025/internal/models"
	"github.com/mastastny/slavoj-web-2025/internal/repository"
	"github.com/mastastny/slavoj-web-2025/internal/service"
)

var echoLambda *echoadapter.EchoLambdaV2

func main() {
	conf := config.NewConfig()

	db := database.Init()
	defer db.Close()
	server := handlers.NewServer(db)

	courts := []models.Court{
		{ID: 1, Name: "multifunkční hřiště"},
		{ID: 2, Name: "tenisový kurt č. 1 (antuka)"},
		{ID: 3, Name: "tenisový kurt č. 2 (antuka)"},
		{ID: 4, Name: "volejbalové kurty"},
		{ID: 5, Name: "tartanová dráha"},
		{ID: 6, Name: "travnaté hřiště"},
	}

	eventRepository := repository.NewEventRepository(db)
	reservationService := service.ConstructReservation(eventRepository)
	reservationHandler := handlers.Construct(reservationService, courts)

	e := echo.New()
	e.Static("/", "static")

	e.GET("/", handlers.GetHome)
	e.GET("/about", handlers.GetAbout)
	e.GET("/areals", handlers.GetAreals)
	e.GET("/reservation", reservationHandler.GetReservation)
	e.GET("/membership", handlers.GetMembership)
	e.GET("/contacts", handlers.GetContacts)
	e.GET("/modal", handlers.GetModal)
	e.GET("/documents", handlers.GetDocuments)
	e.GET("/home", handlers.GetHomeContent)

	e.GET("/api/events", server.GetEvents)

	e.POST("/api/reservation", reservationHandler.PostReservation)
	e.GET("/api/reservation/form", handlers.GetModalBookingForm)

	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		echoLambda = echoadapter.NewV2(e)
		lambda.Start(handler)
	} else {
		e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", conf.Port)))
	}
}

func handler(ctx context.Context, req events.APIGatewayV2HTTPRequest) (events.APIGatewayV2HTTPResponse, error) {
	return echoLambda.ProxyWithContext(ctx, req)
}
