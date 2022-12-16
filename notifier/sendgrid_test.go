package notifier

import (
	"testing"

	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

func Test_SendGrid(t *testing.T) {
	base := &Base{
		viper: viper.New(),
	}

	base.viper.Set("from", "from@myhost.com")
	base.viper.Set("to", "to@myhost.com")
	base.viper.Set("token", "this-is-token")

	s := NewSendGrid(base)
	assert.Equal(t, "SendGrid", s.Service)
	assert.Equal(t, "POST", s.method)
	assert.Equal(t, "application/json", s.contentType)

	url, err := s.buildWebhookURL("")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.sendgrid.com/v3/mail/send", url)

	body, err := s.buildBody("This is title", "This is body")
	assert.NoError(t, err)
	assert.Equal(t, `{"personalizations":[{"to":[{"email":"to@myhost.com","name":"to@myhost.com"}]}],"from":{"email":"from@myhost.com","name":"from@myhost.com"},"subject":"This is title","content":[{"type":"text/plain","value":"This is body"}]}`, string(body))

	respBody := `{"errors":[{"message":"Invalid type. Expected: array, given: object.","field":"content","help":"http://sendgrid.com/docs/API_Reference/Web_API_v3/Mail/errors.html#message.content"}]}`
	err = s.checkResult(200, []byte(respBody))
	assert.NoError(t, err)

	err = s.checkResult(202, []byte(respBody))
	assert.NoError(t, err)

	err = s.checkResult(403, []byte(respBody))
	assert.EqualError(t, err, "status: 403, body: "+respBody)

	headers := s.buildHeaders()
	assert.Equal(t, "Bearer this-is-token", headers["Authorization"])
}
