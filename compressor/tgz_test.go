package compressor

import (
	"github.com/huacnlee/gobackup/helper"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTgz_options(t *testing.T) {
	ctx := &Tgz{}
	opts := ctx.options()
	if helper.IsGnuTar {
		assert.Equal(t, opts[0], "--ignore-failed-read")
		assert.Equal(t, opts[1], "-zcf")
	} else {
		assert.Equal(t, opts[0], "-zcf")
	}

}
