package main

import (
	"context"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/joho/godotenv/autoload"
	"log"
	"os"
	bot2 "presentation_pool/pkg/bot"
	"presentation_pool/pkg/excel"
)

var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonURL("1.com", "http://1.com"),
		tgbotapi.NewInlineKeyboardButtonData("2", "2"),
		tgbotapi.NewInlineKeyboardButtonData("3", "3"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("4", "4"),
		tgbotapi.NewInlineKeyboardButtonData("5", "5"),
		tgbotapi.NewInlineKeyboardButtonData("6", "6"),
	),
)

func main() {
	cr := []byte(os.Getenv("SERVICE_ACCOUNT_CREDENTIALS"))
	store, err := excel.New(context.Background(), "1s2EWBJUPQPYfX97oUfHfwZSasUnKXbDR7dcB-gu3wzE", cr)
	if err != nil {
		log.Fatalf("Unable to retrieve Sheets client: %v", err)
	}

	service, err := bot2.New(os.Getenv("TELEGRAM_APITOKEN"), store)
	if err != nil {
		log.Fatalf("start bot: %v", err)
	}

	service.Run()
}
