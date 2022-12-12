package notifier

import (
	"encoding/json"
	"fmt"
)

type discordPayload struct {
	Content string `json:"content"`
}

func NewDiscord(base *Base) *Webhook {
	return &Webhook{
		Base:        *base,
		Service:     "feishu",
		method:      "POST",
		contentType: "application/json",
		buildBody: func(title, message string) ([]byte, error) {
			payload := discordPayload{
				Content: fmt.Sprintf("%s\n\n%s", title, message),
			}

			return json.Marshal(payload)
		},
	}
}
