package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"presentation_pool/pkg/models"
	"strings"
)

func (b *Bot) adminHandler(req tgbotapi.Update, user *models.User) (tgbotapi.Chattable, error) {
	if req.Message != nil && req.Message.IsCommand() { // ignore any non-command Messages
		return b.adminCommand(req)
	}

	return b.adminCallback(req, user)
}

func (b *Bot) adminCallback(req tgbotapi.Update, user *models.User) (tgbotapi.Chattable, error) {
	if req.CallbackQuery == nil {
		return nil, ErrNotAllow
	}

	cb, ok := extractAdminCallback(req.CallbackData())
	if !ok {
		return b.userHandlerCallback(req, user)
	}

	switch b.status.Status {
	case models.StatusInProgress:
		return tgbotapi.NewMessage(req.FromChat().ID, ""), nil
	//case models.StatusComplete:
	//	return tgbotapi.NewMessage(req.FromChat().ID, "")
	default:
		return b.adminHandleVotes(cb, req), nil
	}
}

func (b *Bot) adminCommand(req tgbotapi.Update) (tgbotapi.Chattable, error) {
	switch req.Message.Command() {
	case "begin":
		return b.msgShowAdminVotesWindow(req.FromChat().ID), nil
	case "complete":
		if err := b.CompleteStep(); err != nil {
			return tgbotapi.NewMessage(req.FromChat().ID, err.Error()), nil
		}

		return b.msgShowStatus(req.FromChat().ID), nil
	case "next":
		if err := b.NextStep(); err != nil {
			return tgbotapi.NewMessage(req.FromChat().ID, err.Error()), nil
		}

		return b.msgShowStatus(req.FromChat().ID), nil
	case "status":
		return b.msgShowStatus(req.FromChat().ID), nil
	case "show":
		return tgbotapi.NewMessage(req.FromChat().ID, ""), nil
	case "reload":
		b.Load()
		return tgbotapi.NewMessage(req.FromChat().ID, "OK"), nil
	default:
		return b.userCommand(req)
	}
}

func (b *Bot) adminHandleVotes(sel string, req tgbotapi.Update) tgbotapi.MessageConfig {
	for _, vote := range b.votes {
		if vote.Name != sel {
			continue
		}

		if err := b.StartStep(&vote); err != nil {
			return tgbotapi.NewMessage(req.FromChat().ID, fmt.Sprintf("error: %v", err))
		}

		return tgbotapi.NewMessage(req.FromChat().ID, fmt.Sprintf("start: %s", sel))
	}

	return tgbotapi.NewMessage(req.FromChat().ID, "no votes settings")
}

const (
	CommandShowVotes = "Select vote"
)

// adminCallbackPrepare help select callback for admin or for others
func adminCallbackPrepare(name string) string {
	return fmt.Sprintf("admin:%s", name)
}

func extractAdminCallback(name string) (string, bool) {
	if !strings.HasPrefix(name, "admin:") {
		return "", false
	}

	return strings.TrimPrefix(name, "admin:"), true
}

func (b *Bot) msgShowAdminVotesWindow(chatID int64) tgbotapi.MessageConfig {
	if len(b.votes) == 0 {
		return tgbotapi.NewMessage(chatID, "no votes settings")
	}

	var x []tgbotapi.InlineKeyboardButton
	for _, vote := range b.votes {
		x = append(x, tgbotapi.NewInlineKeyboardButtonData(vote.Name, adminCallbackPrepare(vote.Name)))
	}

	var xc = tgbotapi.NewInlineKeyboardMarkup(x)
	msg := tgbotapi.NewMessage(chatID, CommandShowVotes)
	msg.ReplyMarkup = xc

	return msg
}

func (b *Bot) msgShowStatus(chatID int64) tgbotapi.MessageConfig {
	msg := fmt.Sprintf("%s/%s => %d", b.status.VoteName, b.status.Status, b.status.Step)
	return tgbotapi.NewMessage(chatID, msg)
}
