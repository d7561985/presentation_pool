package api

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"log"
	"presentation_pool/pkg/bot"
)

type Handler interface {
	Handle(update tgbotapi.Update) (tgbotapi.Chattable, error)
}

type API struct {
	api *tgbotapi.BotAPI
	bot *bot.Bot
}

func New(Token string, bot *bot.Bot) (*API, error) {
	a, err := tgbotapi.NewBotAPI(Token)
	if err != nil {
		return nil, errors.WithMessagef(err, "creation bot api")
	}

	a.Debug = true

	log.Printf("Authorized on account %s", a.Self.UserName)

	srv := &API{
		api: a,
		bot: bot,
	}

	return srv, err
}

func (b *API) Run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		if update.CallbackQuery != nil {
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			if _, err := b.api.Request(callback); err != nil {
				log.Printf("ERR/send: %v", err)
				continue
			}
		}

		msg, err := b.bot.Handle(update)
		if err != nil {
			log.Printf("ERR/handle: %v", err)
			continue
		}

		if _, err = b.api.Send(msg); err != nil {
			log.Printf("ERR/send: %v", err)
			continue
		}

		if update.CallbackQuery != nil {
			m := tgbotapi.NewDeleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID)
			if _, err := b.api.Request(m); err != nil {
				log.Printf("ERR/send: %v", err)
				continue
			}
		}
	}
}

func (b *API) BroadcastWorker(msgs <-chan tgbotapi.MessageConfig) {
	for msg := range msgs {
		if _, err := b.api.Send(msg); err != nil {
			fmt.Printf("send %v", errors.WithStack(err))
		}
	}
}
