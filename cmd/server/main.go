package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	echoadapter "github.com/awslabs/aws-lambda-go-api-proxy/echo"
	"github.com/labstack/echo/v4"
	config "github.com/mastastny/slavoj-web-2025/internal/config"
	"github.com/mastastny/slavoj-web-2025/internal/handlers"
	"github.com/mastastny/slavoj-web-2025/internal/models"
	"github.com/mastastny/slavoj-web-2025/internal/repository"
	"github.com/mastastny/slavoj-web-2025/internal/service"
	"github.com/mastastny/slavoj-web-2025/internal/service/resendEmail"
)

var echoLambda *echoadapter.EchoLambdaV2
var lockerSvc *service.LockerService

func main() {
	conf := config.NewConfig()

	logLevel := slog.LevelInfo
	if conf.LogDebug {
		logLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: logLevel})))

	courts := []models.Court{
		{ID: 1, Name: "multifunkční hřiště"},
		{ID: 2, Name: "tenisový kurt č. 1 (antuka)"},
		{ID: 3, Name: "tenisový kurt č. 2 (antuka)"},
		{ID: 4, Name: "volejbalové kurty"},
		{ID: 5, Name: "tartanová dráha"},
		{ID: 6, Name: "travnaté hřiště"},
	}

	var eventRepository repository.EventRepository
	if conf.Supabase.DatabaseURL != "" {
		repo, err := repository.NewSupabaseEventRepository(conf.Supabase.DatabaseURL, courts)
		if err != nil {
			slog.Warn("supabase repository init failed, falling back to sqlite", "err", err)
			eventRepository = repository.NewSqliteEventRepository(courts)
		} else {
			eventRepository = repo
			slog.Info("supabase event repository created")
		}
	} else {
		eventRepository = repository.NewSqliteEventRepository(courts)
		slog.Info("sqlite event repository created")
	}

	//eventRepository = repository.NewSqliteEventRepository(courts)

	lockerSvc = service.NewLockerService(conf)
	lockerService := lockerSvc
	linkCoder := service.NewLinkCoder(conf)
	emailService := resendEmail.NewEmail(conf)
	reservationService := service.ConstructReservation(eventRepository, emailService, lockerService, linkCoder, conf.PublicDomain)
	reservationHandler := handlers.Construct(reservationService, courts)
	server := handlers.NewServer(eventRepository)

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
	e.GET("/novinky/1", handlers.GetNovinka1)
	e.GET("/email", handlers.GetEmailPreview)
	e.GET("/filter", func(c echo.Context) error {
		err := lockerService.FilterExpiredPasscodesFromLock()
		if err != nil {
			return c.String(500, err.Error())
		}
		return c.String(200, "ok")
	})

	e.GET("/api/events", server.GetEvents)

	e.POST("/api/reservation", reservationHandler.PostReservation)
	e.GET("/api/reservation/form", handlers.GetModalBookingForm)
	e.GET("/reservation/cancel/:encodedId", reservationHandler.DeleteReservation)

	if os.Getenv("AWS_LAMBDA_FUNCTION_NAME") != "" {
		echoLambda = echoadapter.NewV2(e)
		lambda.Start(handler)
	} else {
		e.Logger.Fatal(e.Start(fmt.Sprintf(":%d", conf.Port)))
	}
}

type schedulerEvent struct {
	Source string `json:"source"`
}

func handler(ctx context.Context, raw json.RawMessage) (any, error) {
	var evt schedulerEvent
	if err := json.Unmarshal(raw, &evt); err == nil && evt.Source == "aws.scheduler" {
		slog.Info("EventBridge scheduler trigger received, running filter")
		err = lockerSvc.FilterExpiredPasscodesFromLock()
		if err != nil {
			slog.Warn("filter failed", "err", err)
		}
		return nil, err
	}

	var req events.APIGatewayV2HTTPRequest
	if err := json.Unmarshal(raw, &req); err != nil {
		return nil, fmt.Errorf("unknown event type: %w", err)
	}
	return echoLambda.ProxyWithContext(ctx, req)
}
