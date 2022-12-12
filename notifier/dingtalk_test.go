package notifier

import (
	"testing"

	"github.com/longbridgeapp/assert"
)

func Test_Dingtalk(t *testing.T) {
	base := &Base{}

	s := NewDingtalk(base)
	assert.Equal(t, "DingTalk", s.Service)
	assert.Equal(t, "POST", s.method)
	assert.Equal(t, "application/json", s.contentType)

	body, err := s.buildBody("This is title", "This is body")
	assert.NoError(t, err)
	assert.Equal(t, `{"msgtype":"text","text":{"content":"This is title\n\nThis is body"}}`, string(body))
}
