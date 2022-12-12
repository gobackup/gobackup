package notifier

import (
	"fmt"
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

	err = s.checkResult(200, []byte(`{"errcode":0,"errmsg":"ok"}`))
	assert.NoError(t, err)

	respBody := `{"errcode":300001,"errmsg":"Invalid token"}`
	err = s.checkResult(403, []byte(respBody))
	assert.EqualError(t, err, fmt.Sprintf("status: %d, body: %s", 403, respBody))
}
