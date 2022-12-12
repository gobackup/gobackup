package notifier

import (
	"encoding/json"
	"fmt"
)

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
		Service:     "feishu",
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
	}
}
