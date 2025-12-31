package notifier

import (
	"testing"

	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

func Test_Telegram(t *testing.T) {
	base := &Base{
		viper: viper.New(),
	}

	s := NewTelegram(base)
	s.viper.Set("token", "123213:this-is-my-token")
	s.viper.Set("chat_id", "@gobackuptest")

	assert.Equal(t, "Telegram", s.Service)
	assert.Equal(t, "POST", s.method)
	assert.Equal(t, "application/json", s.contentType)

	body, err := s.buildBody("This is title", "This is body")
	assert.NoError(t, err)
	assert.Equal(t, `{"chat_id":"@gobackuptest","text":"This is title\n\nThis is body","message_thread_id":""}`, string(body))

	url, err := s.buildWebhookURL("")
	assert.NoError(t, err)
	assert.Equal(t, "https://api.telegram.org/bot123213:this-is-my-token/sendMessage", url)

	respBody := `{"ok":true,"result":{"message_id":1,"from":{"id":123213,"is_bot":true,"first_name":"GoBackup","username":"GoBackupBot"},"chat":{"id":-100123213,"title":"GoBackup Test","type":"supergroup"},"date":1610000000,"text":"This is title\n\nThis is body"}}`
	err = s.checkResult(200, []byte(respBody))
	assert.NoError(t, err)

	respBody = `{"ok":false,"error_code":403,"description":"Forbidden: bot was blocked by the user"}`
	err = s.checkResult(403, []byte(respBody))
	assert.EqualError(t, err, "status: 403, body: "+respBody)
}
