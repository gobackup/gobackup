package notifier

import (
	"encoding/json"
	"fmt"

	"github.com/gobackup/gobackup/helper"
)

type telegramPayload struct {
	ChatID 		string `json:"chat_id"`
	Text   		string `json:"text"`
	MessageThreadId string `json:"message_thread_id"`
}

const DEFAULT_TELEGRAM_ENDPOINT = "api.telegram.org"

func NewTelegram(base *Base) *Webhook {
	return &Webhook{
		Base:        *base,
		Service:     "Telegram",
		method:      "POST",
		contentType: "application/json",
		buildWebhookURL: func(url string) (string, error) {
			token := base.viper.GetString("token")
			endpoint := DEFAULT_TELEGRAM_ENDPOINT
			if base.viper.IsSet("endpoint") {
				endpoint = base.viper.GetString("endpoint")
			}

			endpoint = helper.FormatEndpoint(endpoint)

			return fmt.Sprintf("%s/bot%s/sendMessage", endpoint, token), nil
		},
		buildBody: func(title, message string) ([]byte, error) {
			chat_id := base.viper.GetString("chat_id")
			message_thread_id := base.viper.GetString("message_thread_id")

			payload := telegramPayload{
				ChatID: chat_id,
				Text:   fmt.Sprintf("%s\n\n%s", title, message),
				MessageThreadId: message_thread_id,
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
