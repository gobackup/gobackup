package compressor

import (
	"testing"

	"github.com/gobackup/gobackup/config"
	"github.com/gobackup/gobackup/helper"
	"github.com/longbridgeapp/assert"
	"github.com/spf13/viper"
)

func TestTar_options(t *testing.T) {
	viper := viper.New()
	viper.Set("args", "--foo --bar --dar")
	base := newBase(config.ModelConfig{
		CompressWith: config.SubConfig{
			Type:  "compress_with",
			Name:  "tar",
			Viper: viper,
		},
	},
	)

	tar := &Tar{base}
	opts := tar.options()
	if helper.IsGnuTar {
		assert.Equal(t, opts[0], "--ignore-failed-read")
		assert.Equal(t, opts[1], "-a")
		assert.Equal(t, opts[2], "-cf")
	} else {
		assert.Equal(t, opts[0], "-a")
		assert.Equal(t, opts[1], "-cf")
		assert.Equal(t, opts[2], "--foo --bar --dar")
	}

}
