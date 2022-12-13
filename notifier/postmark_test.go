package notifier

import (
	"testing"

	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

func Test_Postmark(t *testing.T) {
	base := &Base{
		viper: viper.New(),
	}

	base.viper.Set("from", "from@myhost.com")
	base.viper.Set("to", "to@myhost.com")
	base.viper.Set("token", "this-is-token")

	s := NewPostmark(base)
	assert.Equal(t, "Postmark", s.Service)
	assert.Equal(t, "POST", s.method)
	assert.Equal(t, "application/json", s.contentType)

	url, err := s.buildWebhookURL("")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.postmarkapp.com/email", url)

	body, err := s.buildBody("This is title", "This is body")
	assert.NoError(t, err)
	assert.Equal(t, `{"From":"from@myhost.com","To":"to@myhost.com","Subject":"This is title","TextBody":"This is body","MessageStream":"outbound"}`, string(body))

	err = s.checkResult(200, []byte(`{"ErrorCode":0,"Message":"OK"}`))
	assert.NoError(t, err)

	respBody := `{"ErrorCode":300,"Message":"Invalid token"}`
	err = s.checkResult(403, []byte(respBody))
	assert.EqualError(t, err, "status: 403, body: "+respBody)

	err = s.checkResult(200, []byte(respBody))
	assert.EqualError(t, err, "status: 200, body: "+respBody)

	headers := s.buildHeaders()
	assert.Equal(t, "this-is-token", headers["X-Postmark-Server-Token"])
}
