package notifier

import (
	"testing"

	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

func Test_SES(t *testing.T) {
	base := &Base{
		viper: viper.New(),
	}
	base.viper.Set("access_key_id", "this-is-acces-key-id")
	base.viper.Set("secret_access_key", "this-is-secret-access-key")
	base.viper.Set("token", "this-is-token")

	s := NewSES(base)
	assert.Equal(t, "us-east-1", *s.client.Config.Region)
	credential, err := s.client.Config.Credentials.Get()
	assert.NoError(t, err)
	assert.Equal(t, "this-is-acces-key-id", credential.AccessKeyID)
	assert.Equal(t, "this-is-secret-access-key", credential.SecretAccessKey)
	assert.Equal(t, "this-is-token", credential.SessionToken)
	assert.NotNil(t, s.client)

	base.viper.Set("region", "us-east-2")
	s = NewSES(base)
	assert.Equal(t, "us-east-2", *s.client.Config.Region)
}

func Test_SES_buildEmail(t *testing.T) {
	base := &Base{
		viper: viper.New(),
	}
	base.viper.Set("from", "from@myhost.com")
	base.viper.Set("to", "to@myhost.com")

	s := NewSES(base)
	emailInfo := s.buildEmail("This is title", "This is message")
	assert.Equal(t, "This is title", *emailInfo.Message.Subject.Data)
	assert.Equal(t, "This is message", *emailInfo.Message.Body.Text.Data)
	assert.Equal(t, "from@myhost.com", *emailInfo.Source)
	assert.Equal(t, "to@myhost.com", *emailInfo.Destination.ToAddresses[0])
}
