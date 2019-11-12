package helper

import (
	"runtime"
	"testing"

	"github.com/stretchr/testify/assert"
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
