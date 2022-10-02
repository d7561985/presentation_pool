package bot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"presentation_pool/pkg/models"
)

var (
	ErrNotAllow = errors.New("not allowed")
	ErrInternal = errors.New("internal issue")
)

func (b *Bot) userHandlerMsg(req tgbotapi.Update, user *models.User) (tgbotapi.MessageConfig, error) {
	if b.status.Status != models.StatusInProgress {
		return tgbotapi.NewMessage(req.FromChat().ID, "vote doesn't in progress"), nil
	}

	// command
	if req.Message != nil && req.Message.IsCommand() { // ignore any non-command Messages
		switch req.Message.Command() {
		case "show", "start":
			return b.msgShowCurrentStepWindow(req.FromChat().ID)
		default:
			return tgbotapi.MessageConfig{}, ErrNotAllow
		}
	}

	txt, err := b.userHandlerCallback(req, user)
	if err != nil {
		return tgbotapi.MessageConfig{}, err
	}

	return tgbotapi.NewMessage(req.FromChat().ID, txt), nil
}

func (b *Bot) userHandlerCallback(req tgbotapi.Update, user *models.User) (string, error) {
	if req.CallbackQuery == nil {
		return "", ErrNotAllow
	}

	if b.vote == nil {
		return "", errors.WithStack(ErrInternal)
	}

	var step = b.vote.Steps[b.status.Step]

	if req.CallbackQuery.Message.Text != step.Question {
		return "wrong question", nil
	}

	var ok bool

	for _, s := range step.Option {
		if s == req.CallbackData() {
			ok = true
		}
	}

	if !ok {
		return "cant find option", nil
	}

	if err := b.store.SaveUserVote(b.vote.Name, b.status.Step, step.Question, req.CallbackData(), user); err != nil {
		return "", errors.WithStack(err)
	}

	return "Thank you for your answering!!!", nil

}
