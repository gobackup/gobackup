package notifier

import (
	"encoding/json"
	"fmt"
)

// resendPayload
// https://resend.com/docs/api-reference/emails/send-email
type resendPayload struct {
	To      []string `json:"to"`
	From    string   `json:"from"`
	Subject string   `json:"subject"`
	Body    string   `json:"text"`
}

type resendResponse struct {
	Message string `json:"message"`
	Code    int    `json:"statusCode"`
}

// NewResend
//
// type: resend
// from: from@your-host.com
// to: your-email@xxx.com
// token: xxxxxxxxxxxxxxxxxxx
func NewResend(base *Base) *Webhook {
	return &Webhook{
		Base:        *base,
		Service:     "Resend",
		method:      "POST",
		contentType: "application/json",
		buildWebhookURL: func(url string) (string, error) {
			return "https://api.resend.com/emails", nil
		},
		buildBody: func(title, message string) ([]byte, error) {
			payload := resendPayload{
				From: base.viper.GetString("from"),
				To: []string{
					base.viper.GetString("to"),
				},
				Subject: title,
				Body:    message,
			}

			return json.Marshal(payload)
		},
		buildHeaders: func() map[string]string {
			return map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", base.viper.GetString("token")),
			}
		},
		checkResult: func(status int, body []byte) error {
			if status < 300 {
				return nil
			}

			resp := resendResponse{}
			err := json.Unmarshal(body, &resp)
			if err != nil {
				return fmt.Errorf("status: %d, body: %s", status, string(body))
			}

			return fmt.Errorf("status: %d, message: %s", status, resp.Message)
		},
	}
}
