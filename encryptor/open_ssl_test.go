package encryptor

import (
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestOpenSSL_options(t *testing.T) {
	ctx := &OpenSSL{
		password: "foo(872",
		salt:     false,
		base64:   true,
	}

	opts := strings.Join(ctx.options(), " ")
	assert.Equal(t, opts, "aes-256-cbc -base64 -k foo(872")

	ctx.salt = true
	opts = strings.Join(ctx.options(), " ")
	assert.Equal(t, opts, "aes-256-cbc -base64 -salt -k foo(872")

	ctx.base64 = false
	opts = strings.Join(ctx.options(), " ")
	assert.Equal(t, opts, "aes-256-cbc -salt -k foo(872")
}
