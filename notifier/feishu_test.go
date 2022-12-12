package notifier

import (
	"testing"

	"github.com/longbridgeapp/assert"
)

func Test_Feishu(t *testing.T) {
	base := &Base{}

	s := NewFeishu(base)
	assert.Equal(t, "Feishu", s.Service)
	assert.Equal(t, "POST", s.method)
	assert.Equal(t, "application/json", s.contentType)

	body, err := s.buildBody("This is title", "This is body")
	assert.NoError(t, err)
	assert.Equal(t, `{"msg_type":"text","content":{"text":"This is title\n\nThis is body"}}`, string(body))

	respBody := `{"StatusCode":0,"StatusMessage":"success"}`
	err = s.checkResult(200, []byte(respBody))
	assert.NoError(t, err)

	respBody = `{"StatusCode":1000,"StatusMessage":"invalid token"}`
	err = s.checkResult(403, []byte(respBody))
	assert.EqualError(t, err, "status: 403, body: "+respBody)

	err = s.checkResult(200, []byte(respBody))
	assert.NoError(t, err)
}
