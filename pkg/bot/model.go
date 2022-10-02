package bot

import (
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"presentation_pool/pkg/models"
)

func ToUser(in *tgbotapi.User, email string) *models.User {
	return &models.User{
		ID:           fmt.Sprintf("%v", in.ID),
		Email:        email,
		IsBot:        fmt.Sprintf("%t", in.IsBot),
		FirstName:    in.FirstName,
		LastName:     in.LastName,
		UserName:     in.UserName,
		LanguageCode: in.LanguageCode,
	}
}
