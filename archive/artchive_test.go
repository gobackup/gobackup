package archive

import (
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/helper"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestRun(t *testing.T) {
	// with nil Archive
	model := config.ModelConfig{
		Archive: nil,
	}
	err := Run(model)
	assert.NoError(t, err)
}

func TestOptions(t *testing.T) {
	includes := []string{
		"/foo/bar/dar",
		"/bar/foo",
		"/ddd",
	}

	excludes := []string{
		"/hello/world",
		"/cc/111",
	}

	dumpPath := "~/work/dir"

	opts := options(dumpPath, excludes, includes)
	cmd := strings.Join(opts, " ")
	if helper.IsGnuTar {
		assert.Equal(t, cmd, "--ignore-failed-read -cPf ~/work/dir/archive.tar --exclude=/hello/world --exclude=/cc/111 /foo/bar/dar /bar/foo /ddd")
	} else {
		assert.Equal(t, cmd, "-cPf ~/work/dir/archive.tar --exclude=/hello/world --exclude=/cc/111 /foo/bar/dar /bar/foo /ddd")
	}
}
