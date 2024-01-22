package helper

import (
	"runtime"
	"testing"

	"github.com/longbridgeapp/assert"
)

func TestUtils_init(t *testing.T) {
	if runtime.GOOS == "linux" {
		assert.Equal(t, IsGnuTar, true)
	} else {
		assert.Equal(t, IsGnuTar, false)
	}
}

func TestCleanHost(t *testing.T) {
	assert.Equal(t, "foo.bar.com", CleanHost("foo.bar.com"))
	assert.Equal(t, "foo.bar.com", CleanHost("ftp://foo.bar.com"))
	assert.Equal(t, "foo.bar.com", CleanHost("http://foo.bar.com"))
	assert.Equal(t, "", CleanHost("http://"))
}

func TestFormatEndpoint(t *testing.T) {
	assert.Equal(t, "http://foo.bar.com", FormatEndpoint("http://foo.bar.com"))
	assert.Equal(t, "https://foo.bar.com", FormatEndpoint("https://foo.bar.com"))
	assert.Equal(t, "https://foo.bar.com", FormatEndpoint("https://foo.bar.com"))
}
