package notifier

import (
	"encoding/json"
	"fmt"
)

type dingtalkResult struct {
	Code    int    `json:"errcode"`
	Message string `json:"errmsg"`
}

type dingtalkPayload struct {
	MsgType string              `json:"msgtype"`
	Text    dingtalkPayloadText `json:"text"`
}

type dingtalkPayloadText struct {
	Content string `json:"content"`
}

func NewDingtalk(base *Base) *Webhook {
	return &Webhook{
		Base:        *base,
		Service:     "DingTalk",
		method:      "POST",
		contentType: "application/json",
		buildBody: func(title, message string) ([]byte, error) {
			payload := dingtalkPayload{
				MsgType: "text",
				Text: dingtalkPayloadText{
					Content: fmt.Sprintf("%s\n\n%s", title, message),
				},
			}

			return json.Marshal(payload)
		},
		checkResult: func(status int, body []byte) error {
			var result dingtalkResult
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
