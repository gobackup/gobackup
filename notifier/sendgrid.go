package notifier

import (
	"encoding/json"
	"fmt"
)

// sendgridPayload
// https://docs.sendgrid.com/api-reference/mail-send/mail-send
type sendgridPayload struct {
	Personalizations []map[string]sendgridPersonalization `json:"personalizations"`
	From             sendgridEmailInfo                    `json:"from"`
	Subject          string                               `json:"subject"`
	Body             []sendgridContent                    `json:"content"`
}

type sendgridPersonalization = []sendgridEmailInfo
type sendgridEmailInfo struct {
	Email string `json:"email"`
	Name  string `json:"name"`
}
type sendgridContent struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// NewSendGrid
//
// type: sendgrid
// from: from@your-host.com
// to: your-email@xxx.com
// token: xxxxxxxxxxxxxxxxxxx
func NewSendGrid(base *Base) *Webhook {
	return &Webhook{
		Base:        *base,
		Service:     "SendGrid",
		method:      "POST",
		contentType: "application/json",
		buildWebhookURL: func(url string) (string, error) {
			return "https://api.sendgrid.com/v3/mail/send", nil
		},
		buildBody: func(title, message string) ([]byte, error) {
			payload := sendgridPayload{
				From: sendgridEmailInfo{
					Name:  base.viper.GetString("from"),
					Email: base.viper.GetString("from"),
				},
				Personalizations: []map[string][]sendgridEmailInfo{
					{
						"to": []sendgridEmailInfo{
							{
								Name:  base.viper.GetString("to"),
								Email: base.viper.GetString("to"),
							},
						},
					},
				},
				Subject: title,
				Body: []sendgridContent{
					{
						Type:  "text/plain",
						Value: message,
					},
				},
			}

			return json.Marshal(payload)
		},
		buildHeaders: func() map[string]string {
			return map[string]string{
				"Authorization": fmt.Sprintf("Bearer %s", base.viper.GetString("token")),
			}
		},
		checkResult: func(status int, body []byte) error {
			if status >= 300 {
				return fmt.Errorf("status: %d, body: %s", status, string(body))
			}

			return nil
		},
	}
}
