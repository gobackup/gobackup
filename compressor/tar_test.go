package compressor

import (
	"testing"

	"github.com/gobackup/gobackup/helper"
	"github.com/longbridgeapp/assert"
)

func TestTar_options(t *testing.T) {
	tar := &Tar{}
	opts := tar.options()
	if helper.IsGnuTar {
		assert.Equal(t, opts[0], "--ignore-failed-read")
		assert.Equal(t, opts[1], "-a")
		assert.Equal(t, opts[2], "-cf")
	} else {
		assert.Equal(t, opts[0], "-a")
		assert.Equal(t, opts[1], "-cf")
	}

}
