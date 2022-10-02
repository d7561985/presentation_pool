package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"log"
	"presentation_pool/pkg/excel"
	"presentation_pool/pkg/models"
	"regexp"
	"strconv"
	"strings"
)

var reg = regexp.MustCompile(`\S@\S+\.\S+`)

func emailValidation(email string) bool {
	return reg.MatchString(email)
}

type Bot struct {
	api   *tgbotapi.BotAPI
	store *excel.Excel

	// in_memory cache
	status *models.StatusData
	vote   *models.Vote

	votes []models.Vote
}

func New(token string, store *excel.Excel) (*Bot, error) {
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		return nil, errors.WithMessagef(err, "creation bot api")
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	status, err := store.GetStatus()
	if err != nil {
		log.Fatalf("cant get status: %v", err)
	}

	votes, err := store.GetAllVotes()
	if err != nil {
		log.Fatalf("cant get votes: %v", err)
	}

	// better change it as seet
	var vote *models.Vote
	if status.Status == models.StatusInProgress {
		for i, v := range votes {
			if v.Name == status.VoteName {
				vote = &votes[i]
			}
		}

		if vote == nil {
			log.Fatalf("cant finde settings for current running vote")
		}
	}

	return &Bot{
		api:    bot,
		store:  store,
		status: status,
		votes:  votes,
		vote:   vote,
	}, nil
}

func (b *Bot) Run() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := b.api.GetUpdatesChan(u)

	for update := range updates {
		user, ok := b.getUser(update)
		if !ok {

		}

		user, err := b.store.GetUser(fmt.Sprintf("%v", update.SentFrom().ID))
		if err != nil {
			/// check if user send email
			if update.Message != nil {
				email := strings.TrimSpace(update.Message.Text)

				if emailValidation(email) {
					u := ToUser(update.SentFrom(), email)
					if err := b.store.SaveUser(u); err != nil {
						log.Println("ERR: save error", err)
					}

					continue
				}
			}

			msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Please enter corporate email")
			b.api.Send(msg)

			continue
		}

		var msg tgbotapi.MessageConfig

		if user.IsAdmin {
			fmt.Println("is admin mode")
			if msg = b.adminHandler(update); msg.Text != "" {
				if _, err = b.api.Send(msg); err != nil {
					panic(err)
				}
			}
		}

		msg, err = b.userHandlerMsg(update, user)
		if err != nil {
			log.Printf("ERR/userHandlerCallback: %v", err)
			continue
		}

		if _, err = b.api.Send(msg); err != nil {
			panic(err)
		}

		// delete selection
		if update.CallbackQuery != nil {
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			if _, err := b.api.Request(callback); err != nil {
				panic(err)
			}

			m := tgbotapi.NewDeleteMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Message.MessageID)
			if _, err = b.api.Request(m); err != nil {
				panic(err)
			}
		}
	}
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

		if _, err = b.api.Send(msg); err != nil {
			log.Println("ERR: Broadcast => send", err)
		}
	}

	return nil
}

func (b *Bot) msgShowCurrentStepWindow(chatID int64) (tgbotapi.MessageConfig, error) {
	if int(b.status.Step) >= len(b.vote.Steps) {
		return tgbotapi.MessageConfig{}, fmt.Errorf("steps overlap")
	}

	var step = b.vote.Steps[b.status.Step]
	var x [][]tgbotapi.InlineKeyboardButton

	for _, s := range step.Option {
		v := tgbotapi.NewInlineKeyboardButtonData(s, s)
		x = append(x, []tgbotapi.InlineKeyboardButton{v})
	}

	var xc = tgbotapi.NewInlineKeyboardMarkup(x...)
	msg := tgbotapi.NewMessage(chatID, step.Question)
	msg.ReplyMarkup = xc

	return msg, nil
}
