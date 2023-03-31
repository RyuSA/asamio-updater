package discord

import (
	"bytes"
	"encoding/json"
	"net/http"
)

var (
	DISCORD_BOT_USER_NAME = "AsamioUpdater"
)

type DiscordPayload struct {
	Username string `json:"username"`
	Content  string `json:"content"`
}

type DiscordWebhook struct {
	Url string
}

func NewDiscordWebhook(url string) *DiscordWebhook {
	return &DiscordWebhook{
		Url: url,
	}
}

func NewDiscordPayload(content string) *DiscordPayload {
	return &DiscordPayload{
		Username: DISCORD_BOT_USER_NAME,
		Content:  content,
	}
}

func (d *DiscordWebhook) Do(payload *DiscordPayload) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	if _, err := http.Post(d.Url, "application/json", bytes.NewBuffer(body)); err != nil {
		return err
	}
	return nil
}
