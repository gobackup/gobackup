package notifier

import (
	"testing"

	"github.com/longbridgeapp/assert"
)

func Test_Feishu(t *testing.T) {
	base := &Base{}

	s := NewFeishu(base)
	assert.Equal(t, "POST", s.method)
	assert.Equal(t, "application/json", s.contentType)

	body, err := s.buildBody("This is title", "This is body")
	assert.NoError(t, err)
	assert.Equal(t, `{"msg_type":"text","content":{"text":"This is title\n\nThis is body"}}`, string(body))
}
