package notifier

import (
	"testing"

	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

func Test_Webhook(t *testing.T) {
	base := &Base{
		viper: viper.New(),
	}

	base.viper.Set("headers", map[string]string{
		"Authorization": "Bearer this-is-token",
	})

	s := NewWebhook(base)
	assert.Equal(t, "Webhook", s.Service)
	assert.Equal(t, "POST", s.method)
	assert.Equal(t, "application/json", s.contentType)

	base.viper.Set("method", "PUT")
	s = NewWebhook(base)
	assert.Equal(t, "PUT", s.method)

	body, err := s.buildBody("This is title", "This is body")
	assert.NoError(t, err)
	assert.Equal(t, `{"title":"This is title","message":"This is body"}`, string(body))

	headers := s.buildHeaders()
	assert.Equal(t, "Bearer this-is-token", headers["Authorization"])

	err = s.checkResult(200, []byte(`{"status":"ok"}`))
	assert.NoError(t, err)

	respBody := `{"status":"error"}`
	err = s.checkResult(403, []byte(respBody))
	assert.EqualError(t, err, "status: 403, body: "+respBody)
}
