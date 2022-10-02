package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"log"
	"presentation_pool/pkg/models"
	"strings"
)

// getUser
// @return bool - when no further processing messages required
func (b *Bot) getUser(req tgbotapi.Update) (*models.User, bool) {
	if req.SentFrom() == nil {
		return nil, false
	}

	id := fmt.Sprintf("%v", req.SentFrom().ID)
	user, err := b.store.GetUser(id)
	if err == nil {
		return user, true
	}

	msg := tgbotapi.NewMessage(req.Message.Chat.ID, "Please enter corporate email")

	defer func() {
		_, _ = b.api.Send(msg)
	}()

	/// check if user send email
	if req.Message != nil {
		email := strings.TrimSpace(req.Message.Text)

		if emailValidation(email) {
			u := ToUser(req.SentFrom(), email)
			if err = b.store.SaveUser(u); err != nil {
				log.Println("ERR: save error", err)
				return nil, false
			}

			// OK
			if b.status.Status == models.StatusInProgress {
				msg, err = b.msgShowCurrentStepWindow(req.Message.Chat.ID)
				_, _ = b.api.Send(msg)
				return nil, false
			}

		}
	}

	return nil, false
}
