package helper

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
)

func TestIsExistsPath(t *testing.T) {
	exist := IsExistsPath("foo/bar")
	assert.False(t, exist)

	exist = IsExistsPath("./filepath_test.go")
	assert.True(t, exist)
}

func TestMkdirP(t *testing.T) {
	dest := path.Join(os.TempDir(), "test-mkdir-p")
	exist := IsExistsPath(dest)
	assert.False(t, exist)

	MkdirP(dest)
	defer os.Remove(dest)
	exist = IsExistsPath(dest)
	assert.True(t, exist)
}
