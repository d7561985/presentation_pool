package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/pkg/errors"
	"presentation_pool/pkg/models"
	"regexp"
	"strings"
)

var reg = regexp.MustCompile(`\S@\S+\.\S+`)

func emailValidation(email string) bool {
	return reg.MatchString(email)
}

func (b *Bot) auth(req tgbotapi.Update) (*models.User, error) {
	id := fmt.Sprintf("%v", req.SentFrom().ID)
	user, err := b.store.GetUser(id)

	return user, errors.WithStack(err)
}

// getUser
// @return bool - when no further processing messages required
func (b *Bot) authHandle(req tgbotapi.Update) (tgbotapi.Chattable, error) {
	/// check if user send email
	if req.Message == nil {
		return tgbotapi.NewMessage(req.FromChat().ID, "Please enter corporate email"), nil
	}

	email := strings.TrimSpace(req.Message.Text)

	if !emailValidation(email) || !strings.Contains(email, b.cfg.AuthRule) {
		return tgbotapi.NewMessage(req.FromChat().ID, fmt.Sprintf("Please enter corporate email [input was: %q]", email)), nil
	}

	u := ToUser(req.SentFrom(), email)
	if err := b.store.SaveUser(u); err != nil {
		return nil, errors.WithStack(err)
	}

	// OK
	if b.status.Status == models.StatusInProgress {
		return b.msgShowCurrentStepWindow(req.FromChat().ID)
	}

	return tgbotapi.NewMessage(req.Message.Chat.ID, "Hi there! Please wait of QUIZ launching."), nil
}
