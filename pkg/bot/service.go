package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"log"
	"presentation_pool/pkg/excel"
	"presentation_pool/pkg/models"
	"strconv"
)

type Bot struct {
	store *excel.Excel

	// in_memory cache
	status *models.StatusData
	vote   *models.Vote

	votes []models.Vote

	cfg Cfg
}

type Cfg struct {

	// AuthRule required contain
	AuthRule string
	Msg      chan<- tgbotapi.MessageConfig
}

func New(cfg Cfg, store *excel.Excel) *Bot {
	srv := &Bot{
		store: store,
		cfg:   cfg,
	}

	srv.Load()

	return srv
}

func (b *Bot) Load() {
	status, err := b.store.GetStatus()
	if err != nil {
		log.Fatalf("cant get status: %v", err)
	}

	votes, err := b.store.GetAllVotes()
	if err != nil {
		log.Fatalf("cant get votes: %v", err)
	}

	// better change it as seet
	var vote *models.Vote
	if status.Status == models.StatusInProgress || status.Status == models.StatusComplete {
		for i, v := range votes {
			if v.Name == status.VoteName {
				vote = &votes[i]
			}
		}

		if vote == nil {
			log.Fatalf("cant finde settings for current running vote")
		}
	}

	b.votes = votes
	b.vote = vote
	b.status = status
}

func (b *Bot) Handle(update tgbotapi.Update) (tgbotapi.Chattable, error) {
	if update.SentFrom() == nil {
		return nil, errors.WithStack(fmt.Errorf("update dont have sendForm"))
	}

	user, err := b.auth(update)
	if err != nil {
		return b.authHandle(update)
	}

	var msg tgbotapi.Chattable

	if user.IsAdmin {
		fmt.Println("is admin mode")
		msg, err = b.adminHandler(update, user)

		return msg, errors.WithStack(err)
	}

	msg, err = b.userHandlerMsg(update, user)

	return msg, errors.WithStack(err)
}

func (b *Bot) Broadcast(msg tgbotapi.MessageConfig) error {
	users, err := b.store.GetUsers()
	if err != nil {
		return errors.WithMessage(err, "cant get users")
	}

	for _, user := range users {
		chatID, err := strconv.ParseInt(user.ID, 10, 64)
		if err != nil {
			log.Println("ERR: Broadcast=> user parse", err)
			continue
		}

		msg.ChatID = chatID
		b.cfg.Msg <- msg
	}

	return nil
}

func (b *Bot) msgShowCurrentStepWindow(chatID int64) (tgbotapi.MessageConfig, error) {
	if int(b.status.Step) >= len(b.vote.Steps) {
		return tgbotapi.MessageConfig{}, fmt.Errorf("steps overlap")
	}

	var step = b.vote.Steps[b.status.Step]
	var x [][]tgbotapi.InlineKeyboardButton

	for id, s := range step.Option {
		if s == "" {
			continue
		}

		v := tgbotapi.NewInlineKeyboardButtonData(s, fmt.Sprintf("%d", id))
		x = append(x, []tgbotapi.InlineKeyboardButton{v})
	}

	var xc = tgbotapi.NewInlineKeyboardMarkup(x...)
	msg := tgbotapi.NewMessage(chatID, step.Question)
	msg.ReplyMarkup = xc

	return msg, nil
}
