package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"presentation_pool/pkg/models"
	"strings"
)

var (
	ErrNotAllow = errors.New("not allowed")
	ErrInternal = errors.New("internal issue")
)

func (b *Bot) userHandlerMsg(req tgbotapi.Update, user *models.User) (tgbotapi.Chattable, error) {
	if b.status.Status != models.StatusInProgress {
		return tgbotapi.NewMessage(req.FromChat().ID, "vote doesn't in progress"), nil
	}

	// command
	if req.Message != nil && req.Message.IsCommand() { // ignore any non-command Messages
		return b.userCommand(req)
	}

	msg, err := b.userHandlerCallback(req, user)
	return msg, errors.WithStack(err)
}

func (b *Bot) userCommand(req tgbotapi.Update) (tgbotapi.Chattable, error) {
	switch req.Message.Command() {
	case "show", "start":
		return b.msgShowCurrentStepWindow(req.FromChat().ID)
	default:
		return tgbotapi.MessageConfig{}, ErrNotAllow
	}
}

func (b *Bot) userHandlerCallback(req tgbotapi.Update, user *models.User) (tgbotapi.Chattable, error) {
	if req.CallbackQuery == nil {
		return nil, ErrNotAllow
	}

	if b.status.Status != models.StatusInProgress {
		return tgbotapi.NewMessage(req.FromChat().ID, "vote doesn't in progress"), nil
	}

	if b.vote == nil || int(b.status.Step) >= len(b.vote.Steps) {
		return nil, errors.WithStack(ErrInternal)
	}

	var step = b.vote.Steps[b.status.Step]

	if req.CallbackQuery.Message.Text != strings.TrimSpace(step.Question) {
		return tgbotapi.NewMessage(req.FromChat().ID, "wrong question"), nil
	}

	var answer string

	for id, a := range step.Option {
		if fmt.Sprintf("%d", id) == req.CallbackData() {
			answer = a
			break
		}
	}

	if answer == "" {
		return tgbotapi.NewMessage(req.FromChat().ID, "cant find option"), nil
	}

	if err := b.store.SaveUserVote(b.vote.Name, b.status.Step, step.Question, answer, user); err != nil {
		return nil, errors.WithStack(err)
	}

	return tgbotapi.NewMessage(req.FromChat().ID, "Thank you for your answering!!!"), nil
}
