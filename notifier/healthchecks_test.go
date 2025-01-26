package notifier

import (
	"testing"

	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

func Test_Healthchecks(t *testing.T) {
	base := &Base{
		viper: viper.New(),
	}

	s := NewHealthchecks(base)
	assert.Equal(t, "Healthchecks", s.Service)

	err := s.checkResult(200, []byte(`{"status":"ok"}`))
	assert.NoError(t, err)

	respBody := `{"status":"error"}`
	err = s.checkResult(403, []byte(respBody))
	assert.EqualError(t, err, "status: 403, body: "+respBody)
}
