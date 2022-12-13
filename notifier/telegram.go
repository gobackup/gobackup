package notifier

import (
	"encoding/json"
	"fmt"
)

type telegramPayload struct {
	ChatID string `json:"chat_id"`
	Text   string `json:"text"`
}

func NewTelegram(base *Base) *Webhook {
	return &Webhook{
		Base:        *base,
		Service:     "Telegram",
		method:      "POST",
		contentType: "application/json",
		buildWebhookURL: func(url string) (string, error) {
			token := base.viper.GetString("token")

			return fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token), nil
		},
		buildBody: func(title, message string) ([]byte, error) {
			chat_id := base.viper.GetString("chat_id")

			payload := telegramPayload{
				ChatID: chat_id,
				Text:   fmt.Sprintf("%s\n\n%s", title, message),
			}

			return json.Marshal(payload)
		},
		checkResult: func(status int, body []byte) error {
			if status != 200 {
				return fmt.Errorf("status: %d, body: %s", status, string(body))
			}

			return nil
		},
	}
}
