package main

import (
	"context"
	_ "github.com/joho/godotenv/autoload"
	"log"
	"os"
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

	cfg := bot2.Cfg{
		Token:    os.Getenv("TELEGRAM_APITOKEN"),
		AuthRule: os.Getenv("EMAIL_PATTERN"),
	}

	service, err := bot2.New(cfg, store)
	if err != nil {
		log.Fatalf("start bot: %v", err)
	}

	service.Run()
}
