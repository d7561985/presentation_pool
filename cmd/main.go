package main

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/joho/godotenv/autoload"
	"log"
	"os"
	"presentation_pool/pkg/api"
	bot2 "presentation_pool/pkg/bot"
	"presentation_pool/pkg/excel"
)

func main() {
	cr := []byte(os.Getenv("SERVICE_ACCOUNT_CREDENTIALS"))
	sheetID := os.Getenv("SHEET_ID")

	store, err := excel.New(context.Background(), sheetID, cr)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	msg := make(chan tgbotapi.MessageConfig, 100)
	defer func() {
		close(msg)
	}()

	controller := bot2.New(bot2.Cfg{
		AuthRule: os.Getenv("EMAIL_PATTERN"),
		Msg:      msg,
	}, store)

	server, err := api.New(os.Getenv("TELEGRAM_APITOKEN"), controller)
	if err != nil {
		log.Fatalf("start api: %v", err)
	}

	go server.BroadcastWorker(msg)

	server.Run()

}
