package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"os"
	"presentation_pool/pkg/api"
	bot2 "presentation_pool/pkg/bot"
	"presentation_pool/pkg/excel"
	"strconv"
)

func main() {
	cr := []byte(os.Getenv("SERVICE_ACCOUNT_CREDENTIALS"))
	sheetID := os.Getenv("SHEET_ID")

	store, err := excel.New(context.Background(), sheetID, cr)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	// should be unbuffered
	msg := make(chan tgbotapi.MessageConfig)
	defer func() {
		close(msg)
	}()

	controller := bot2.New(bot2.Cfg{
		AuthRule: os.Getenv("EMAIL_PATTERN"),
		Msg:      msg,
	}, store)

	// ignore parse bool issue
	isWH, _ := strconv.ParseBool(os.Getenv("IS_WEBHOOK"))

	x := api.Cfg{
		Token: os.Getenv("TELEGRAM_APITOKEN"),
		IsWH:  isWH,
		Host:  os.Getenv("WEBHOOK_ADDR"),
	}

	server, err := api.New(x, controller)
	if err != nil {
		log.Fatalf("start api: %v", err)
	}

	go server.BroadcastWorker(msg)

	lambda.Start(func(ctx context.Context, request events.ALBTargetGroupRequest) (events.ALBTargetGroupResponse, error) {
		fmt.Printf("Processing request data for traceId %s.\n", request.Headers["x-amzn-trace-id"])
		fmt.Printf("Body size = %d.\n", len(request.Body))

		fmt.Println("Headers:", request.Headers)

		var update tgbotapi.Update
		if err = json.Unmarshal([]byte(request.Body), &update); err != nil {
			return events.ALBTargetGroupResponse{Body: request.Body, StatusCode: 500, StatusDescription: "200 OK", IsBase64Encoded: false, Headers: map[string]string{}}, err
		}

		server.Process(update)

		return events.ALBTargetGroupResponse{Body: request.Body, StatusCode: 200, StatusDescription: "200 OK", IsBase64Encoded: false, Headers: map[string]string{}}, nil
	})
}
