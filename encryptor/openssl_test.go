package encryptor

import (
	"strings"
	"testing"

	"github.com/longbridgeapp/assert"
)

func TestOpenSSL_options(t *testing.T) {
	enc := &OpenSSL{
		password: "foo(872",
		salt:     false,
		base64:   true,
	}

	opts := strings.Join(enc.options(), " ")
	assert.Equal(t, opts, "aes-256-cbc -base64 -k foo(872")

	enc.salt = true
	opts = strings.Join(enc.options(), " ")
	assert.Equal(t, opts, "aes-256-cbc -base64 -salt -k foo(872")

	enc.base64 = false
	opts = strings.Join(enc.options(), " ")
	assert.Equal(t, opts, "aes-256-cbc -salt -k foo(872")
}
