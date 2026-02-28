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
)

var echoLambda *echoadapter.EchoLambdaV2

func main() {
	conf := config.NewConfig()

	db := database.Init()
	defer db.Close()
	server := handlers.NewServer(db)

	e := echo.New()
	e.Static("/", "static")

	e.GET("/", handlers.GetHome)
	e.GET("/about", handlers.GetAbout)
	e.GET("/areals", handlers.GetAreals)
	e.GET("/reservation", handlers.GetReservation)
	e.GET("/contacts", handlers.GetContacts)
	e.GET("/modal", handlers.GetModal)
	e.GET("/documents", handlers.GetDocuments)
	e.GET("/home", handlers.GetHomeContent)

	e.GET("/api/events", server.GetEvents)

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
