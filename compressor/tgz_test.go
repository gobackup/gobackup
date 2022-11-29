package compressor

import (
	"testing"

	"github.com/huacnlee/gobackup/helper"
	"github.com/longbridgeapp/assert"
)

func TestTgz_options(t *testing.T) {
	ctx := &Tgz{}
	opts := ctx.options()
	if helper.IsGnuTar {
		assert.Equal(t, "--ignore-failed-read", opts[0])
	}
}
