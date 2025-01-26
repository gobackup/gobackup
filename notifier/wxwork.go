package notifier

import (
	"encoding/json"
	"fmt"
)

type wxworkResult struct {
	Code    int    `json:"errcode"`
	Message string `json:"errmsg"`
}

type wxworkPayload struct {
	MsgType string               `json:"msgtype"`
	Text    wxworkPayloadContent `json:"text"`
}

type wxworkPayloadContent struct {
	Content string `json:"content"`
}

func NewWxWork(base *Base) *Webhook {
	return &Webhook{
		Base:        *base,
		Service:     "WxWork",
		method:      "POST",
		contentType: "application/json",
		buildBody: func(title, message string) ([]byte, error) {
			payload := wxworkPayload{
				MsgType: "text",
				Text: wxworkPayloadContent{
					Content: fmt.Sprintf("%s\n\n%s", title, message),
				},
			}

			return json.Marshal(payload)
		},
		checkResult: func(status int, body []byte) error {
			var result wxworkResult
			err := json.Unmarshal(body, &result)
			if err != nil {
				return err
			}

			if result.Code != 0 || status != 200 {
				return fmt.Errorf("status: %d, body: %s", status, string(body))
			}

			return nil
		},
	}
}
