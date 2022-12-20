package encryptor

import (
	"strings"
	"testing"

	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

func TestOpenSSL_options(t *testing.T) {
	base := &Base{
		viper:       viper.New(),
		archivePath: "/foo/bar",
	}

	enc := NewOpenSSL(base)
	assert.Equal(t, false, enc.base64)
	assert.Equal(t, true, enc.salt)
	assert.Equal(t, "", enc.password)
	assert.Equal(t, "/foo/bar.enc", enc.encryptPath)
	assert.Equal(t, "", enc.args)
	assert.Equal(t, "aes-256-cbc", enc.chiper)
	assert.Equal(t, "aes-256-cbc -salt -k ", strings.Join(enc.options(), " "))

	base.viper.Set("base64", true)
	base.viper.Set("salt", false)
	base.viper.Set("args", "-pbkdf2 -iter 1000")
	base.viper.Set("password", "gobackup-123")
	base.viper.Set("chiper", "rc4")

	enc = NewOpenSSL(base)
	assert.Equal(t, true, enc.base64)
	assert.Equal(t, false, enc.salt)
	assert.Equal(t, "rc4", enc.chiper)
	assert.Equal(t, "gobackup-123", enc.password)
	assert.Equal(t, "-pbkdf2 -iter 1000", enc.args)

	assert.Equal(t, "rc4 -base64 -pbkdf2 -iter 1000 -k gobackup-123", strings.Join(enc.options(), " "))
}
