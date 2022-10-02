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

	service, err := bot2.New(os.Getenv("TELEGRAM_APITOKEN"), store)
	if err != nil {
		log.Fatalf("start bot: %v", err)
	}

	service.Run()
}
