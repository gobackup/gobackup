package notifier

import (
	"encoding/json"
	"fmt"
)

type slackPayload struct {
	Text string `json:"text"`
}

func NewSlack(base *Base) *Webhook {
	return &Webhook{
		Base:        *base,
		Service:     "Slack",
		method:      "POST",
		contentType: "application/json",
		buildBody: func(title, message string) ([]byte, error) {
			payload := slackPayload{
				Text: fmt.Sprintf("%s\n\n%s", title, message),
			}

			return json.Marshal(payload)
		},
	}
}
