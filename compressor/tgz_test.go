package compressor

import (
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/helper"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestTgz(t *testing.T) {
	var ctx Base
	ctx = &Tgz{}
	model := config.ModelConfig{
		Name: "test-tar",
	}
	_, err := ctx.perform(model)
	assert.Error(t, err)
}

func TestTgz_options(t *testing.T) {
	ctx := &Tgz{}
	opts := ctx.options()
	assert.Equal(t, opts[0], "zcf")
	if helper.IsGnuTar {
		assert.Equal(t, opts[1], "--ignore-failed-read")
	}
}
