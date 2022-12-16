package notifier

import (
	"encoding/json"
	"fmt"
)

// postmarkPayload
// https://postmarkapp.com/developer/user-guide/send-email-with-api
type postmarkPayload struct {
	From          string `json:"From"`
	To            string `json:"To"`
	Subject       string `json:"Subject"`
	TextBody      string `json:"TextBody"`
	MessageStream string `json:"MessageStream"`
}

type postmarkResult struct {
	Code    int    `json:"ErrorCode"`
	Message string `json:"Message"`
}

// NewPostmark
//
// type: postmark
// from: from@your-host.com
// to: your-email@xxx.com
// token: xxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx
func NewPostmark(base *Base) *Webhook {
	return &Webhook{
		Base:        *base,
		Service:     "Postmark",
		method:      "POST",
		contentType: "application/json",
		buildWebhookURL: func(url string) (string, error) {
			return "https://api.postmarkapp.com/email", nil
		},
		buildBody: func(title, message string) ([]byte, error) {
			payload := postmarkPayload{
				From:          base.viper.GetString("from"),
				To:            base.viper.GetString("to"),
				Subject:       title,
				TextBody:      message,
				MessageStream: "outbound",
			}

			return json.Marshal(payload)
		},
		buildHeaders: func() map[string]string {
			return map[string]string{
				"X-Postmark-Server-Token": base.viper.GetString("token"),
			}
		},
		checkResult: func(status int, body []byte) error {
			if status != 200 {
				return fmt.Errorf("status: %d, body: %s", status, string(body))
			}

			var result postmarkResult
			if err := json.Unmarshal(body, &result); err != nil {
				return err
			}

			if result.Code != 0 {
				return fmt.Errorf("status: %d, body: %s", status, string(body))
			}

			return nil
		},
	}
}
