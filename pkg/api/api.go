package api

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"log"
	"net/http"
	"presentation_pool/pkg/bot"
	"sync"
)

const (
	port = "8443"
)

type Handler interface {
	Handle(update tgbotapi.Update) (tgbotapi.Chattable, error)
}

type Cfg struct {
	Token string
	IsWH  bool
	Host  string
}

type API struct {
	api *tgbotapi.BotAPI
	bot *bot.Bot

	cfg Cfg

	cb sync.Once
}

func New(cfg Cfg, bot *bot.Bot) (*API, error) {
	a, err := tgbotapi.NewBotAPI(cfg.Token)
	if err != nil {
		return nil, errors.WithMessagef(err, "creation bot api")
	}

	a.Debug = true

	log.Printf("Authorized on account %s", a.Self.UserName)

	srv := &API{
		api: a,
		bot: bot,
		cfg: cfg,
	}

	return srv, err
}

func (b *API) WebhookInit() {
	// "https://www.example.com:8443/"
	adddr := fmt.Sprintf("%s:%s/", b.cfg.Host, port)
	wh, err := tgbotapi.NewWebhook(adddr + b.api.Token)
	if err != nil {
		log.Fatalf("new webhook %v", err)
	}

	_, err = b.api.Request(wh)
	if err != nil {
		log.Fatal(err)
	}

	info, err := b.api.GetWebhookInfo()
	if err != nil {
		log.Fatal(err)
	}

	if info.LastErrorDate != 0 {
		log.Printf("> Telegram callback failed: %s", info.LastErrorMessage)
	}

	go http.ListenAndServe("0.0.0.0:"+port, nil)
}

func (b *API) getUpdate() tgbotapi.UpdatesChannel {
	if b.cfg.IsWH {
		b.cb.Do(func() {
			b.WebhookInit()
		})

		return b.api.ListenForWebhook("/" + b.api.Token)
	}

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	return b.api.GetUpdatesChan(u)
}

func (b *API) Run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.getUpdate()

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
