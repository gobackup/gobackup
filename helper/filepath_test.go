package helper

import (
	"os"
	"path"
	"testing"

	"github.com/longbridgeapp/assert"
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

	assert.Nil(t, MkdirP(dest))
	defer os.Remove(dest)
	exist = IsExistsPath(dest)
	assert.True(t, exist)
}

func TestExplandHome(t *testing.T) {
	newPath := ExplandHome("")
	assert.Equal(t, newPath, "")

	newPath = ExplandHome("/home/jason/111")
	assert.Equal(t, newPath, "/home/jason/111")

	newPath = ExplandHome("~")
	assert.Equal(t, newPath, "~")

	newPath = ExplandHome("~/")
	assert.NotEqual(t, newPath[:2], "~/")

	newPath = ExplandHome("~/foo/bar/dar")
	assert.Equal(t, newPath, path.Join(os.Getenv("HOME"), "/foo/bar/dar"))
}

func TestAbsolutePath(t *testing.T) {
	pwd, _ := os.Getwd()
	newPath := AbsolutePath("foo/bar")
	assert.Equal(t, newPath, path.Join(pwd, "foo/bar"))

	newPath = AbsolutePath("/home/jason/111")
	assert.Equal(t, newPath, "/home/jason/111")

	newPath = AbsolutePath("~")
	assert.NotEqual(t, newPath[:2], "~/")

	newPath = AbsolutePath("~/")
	assert.NotEqual(t, newPath[:2], "~/")

	newPath = AbsolutePath("~/foo/bar/dar")
	assert.Equal(t, newPath, path.Join(os.Getenv("HOME"), "/foo/bar/dar"))
}
