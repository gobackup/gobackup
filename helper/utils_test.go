package helper

import (
	"github.com/stretchr/testify/assert"
	"runtime"
	"testing"
)

func TestUtils_init(t *testing.T) {
	if runtime.GOOS == "linux" {
		assert.Equal(t, IsGnuTar, true)
	} else {
		assert.Equal(t, IsGnuTar, false)
	}
}
