package models

type User struct {
	// ID is a unique identifier for this user or bot
	ID string `json:"id"`

	Email string

	// IsBot true, if this user is a bot
	//
	// optional
	IsBot string `json:"is_bot,omitempty"`

	// FirstName user's or bot's first name
	FirstName string `json:"first_name"`
	// LastName user's or bot's last name
	//
	// optional
	LastName string `json:"last_name,omitempty"`
	// UserName user's or bot's username
	//
	// optional
	UserName string `json:"username,omitempty"`
	// LanguageCode IETF language tag of the user's language
	// more info: https://en.wikipedia.org/wiki/IETF_language_tag
	//
	// optional
	LanguageCode string `json:"language_code,omitempty"`

	IsAdmin bool
}
