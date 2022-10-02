package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"presentation_pool/pkg/models"
)

func (b *Bot) adminHandler(req tgbotapi.Update) tgbotapi.MessageConfig {
	if req.Message != nil && req.Message.IsCommand() { // ignore any non-command Messages
		return b.adminCommand(req)
	}

	if req.CallbackQuery == nil {
		return tgbotapi.NewMessage(req.FromChat().ID, "internal error")
	}

	switch b.status.Status {
	case models.StatusInProgress:
		return tgbotapi.NewMessage(req.FromChat().ID, "")
	case models.StatusComplete:
		return tgbotapi.NewMessage(req.FromChat().ID, "")
	default:
		return b.adminHandleVotes(req)
	}
}

func (b *Bot) adminCommand(req tgbotapi.Update) tgbotapi.MessageConfig {
	switch req.Message.Command() {
	case "start":
		return b.msgShowAdminVotesWindow(req.FromChat().ID)
	case "complete":
		if err := b.CompleteStep(); err != nil {
			return tgbotapi.NewMessage(req.FromChat().ID, err.Error())
		}

		return b.msgShowStatus(req.FromChat().ID)
	case "next":
		if err := b.NextStep(); err != nil {
			return tgbotapi.NewMessage(req.FromChat().ID, err.Error())
		}

		return b.msgShowStatus(req.FromChat().ID)
	case "status":
		return b.msgShowStatus(req.FromChat().ID)
	case "show":
		return tgbotapi.NewMessage(req.FromChat().ID, "")
	case "reload":
		b.Load()
		return tgbotapi.NewMessage(req.FromChat().ID, "OK")
	default:
		return tgbotapi.NewMessage(req.FromChat().ID, "unsupported command: start, complete, next, status, show")
	}
}

func (b *Bot) adminHandleVotes(req tgbotapi.Update) tgbotapi.MessageConfig {
	sel := req.CallbackQuery.Data
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

func (b *Bot) msgShowAdminVotesWindow(chatID int64) tgbotapi.MessageConfig {
	if len(b.votes) == 0 {
		return tgbotapi.NewMessage(chatID, "no votes settings")
	}

	var x []tgbotapi.InlineKeyboardButton
	for _, vote := range b.votes {
		x = append(x, tgbotapi.NewInlineKeyboardButtonData(vote.Name, vote.Name))
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
