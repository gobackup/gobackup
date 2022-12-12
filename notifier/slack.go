package notifier

import (
	"encoding/json"
	"fmt"
)

type slackPayload struct {
	Text string `json:"text"`
}

// type: slack
// url: https://hooks.slack.com/services/QC0UDT3L8/C58B274E5D0/pwFViFlRXon4WWxo30wevzyR
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
		checkResult: func(status int, body []byte) error {
			if status != 200 {
				return fmt.Errorf("status: %d, body: %s", status, string(body))
			}

			return nil
		},
	}
}
