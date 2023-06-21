package notifier

import (
	"testing"

	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

func Test_Resend(t *testing.T) {
	base := &Base{
		viper: viper.New(),
	}

	base.viper.Set("from", "from@myhost.com")
	base.viper.Set("to", "to@myhost.com")
	base.viper.Set("token", "this-is-token")

	s := NewResend(base)
	assert.Equal(t, "Resend", s.Service)
	assert.Equal(t, "POST", s.method)
	assert.Equal(t, "application/json", s.contentType)

	url, err := s.buildWebhookURL("")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.resend.com/emails", url)

	body, err := s.buildBody("This is title", "This is body")
	assert.NoError(t, err)
	assert.Equal(t, `{"to":["to@myhost.com"],"from":"from@myhost.com","subject":"This is title","text":"This is body"}`, string(body))

	respBody := `{"name":"validation_error","message":"This is error message","statusCode":401}`

	err = s.checkResult(200, []byte(respBody))
	assert.NoError(t, err)

	err = s.checkResult(403, []byte(respBody))
	assert.EqualError(t, err, "status: 403, message: This is error message")

	err = s.checkResult(403, []byte("foo bar"))
	assert.EqualError(t, err, "status: 403, body: foo bar")

	headers := s.buildHeaders()
	assert.Equal(t, "Bearer this-is-token", headers["Authorization"])
}
