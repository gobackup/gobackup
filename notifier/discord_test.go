package notifier

import (
	"testing"

	"github.com/longbridgeapp/assert"
)

func Test_Discord(t *testing.T) {
	base := &Base{}

	s := NewDiscord(base)
	assert.Equal(t, "Discord", s.Service)
	assert.Equal(t, "POST", s.method)
	assert.Equal(t, "application/json", s.contentType)

	body, err := s.buildBody("This is title", "This is body")
	assert.NoError(t, err)
	assert.Equal(t, `{"content":"This is title\n\nThis is body"}`, string(body))
}
