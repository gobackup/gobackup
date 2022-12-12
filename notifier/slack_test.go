package notifier

import (
	"testing"

	"github.com/longbridgeapp/assert"
)

func Test_Slack(t *testing.T) {
	base := &Base{}

	s := NewSlack(base)
	assert.Equal(t, "Slack", s.Service)
	assert.Equal(t, "POST", s.method)
	assert.Equal(t, "application/json", s.contentType)

	body, err := s.buildBody("This is title", "This is body")
	assert.NoError(t, err)
	assert.Equal(t, `{"text":"This is title\n\nThis is body"}`, string(body))
}
