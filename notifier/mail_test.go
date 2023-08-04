package notifier

import (
	"testing"

	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

func Test_Mail(t *testing.T) {
	base := Base{
		viper: viper.New(),
	}

	base.viper.Set("username", "user@myhost.com")
	base.viper.Set("to", "to@myhost.com,to1@myhost.com")
	base.viper.Set("password", "this-is-password")
	base.viper.Set("host", "smtp.myhost.com")

	mail, err := NewMail(&base)
	assert.Nil(t, err)

	assert.Equal(t, "user@myhost.com", mail.from)
	assert.Equal(t, []string{"to@myhost.com", "to1@myhost.com"}, mail.to)
	assert.Equal(t, "user@myhost.com", mail.username)
	assert.Equal(t, "this-is-password", mail.password)
	assert.Equal(t, "smtp.myhost.com", mail.host)
	assert.Equal(t, "25", mail.port)

	assert.Equal(t, "smtp.myhost.com:25", mail.getAddr())

	auth := mail.getAuth()
	assert.NotNil(t, auth)

	base.viper.Set("from", "from@myhost.com")
	base.viper.Set("port", "587")

	mail, err = NewMail(&base)
	assert.Nil(t, err)

	assert.Equal(t, "from@myhost.com", mail.from)
	assert.Equal(t, "587", mail.port)
	assert.Equal(t, "smtp.myhost.com:587", mail.getAddr())

	body := mail.buildBody("This is title", "This is body")
	assert.Equal(t, "Content-Transfer-Encoding: base64\nContent-Type: text/plain; charset=\"utf-8\"\nFrom: from@myhost.com\nSubject: This is title\nTo: to@myhost.com,to1@myhost.com\nVGhpcyBpcyBib2R5", body)
}
