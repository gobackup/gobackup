package storage

import (
	"testing"

	"github.com/gobackup/gobackup/config"
	"github.com/longbridgeapp/assert"
)

func TestBase_newBase(t *testing.T) {
	model := config.ModelConfig{}
	archivePath := "/tmp/gobackup/test-storeage/foo.zip"
	s, _ := newBase(model, archivePath, config.SubConfig{})

	assert.Equal(t, s.archivePath, archivePath)
	assert.Equal(t, s.model, model)
	assert.Equal(t, s.viper, model.Viper)
	assert.Equal(t, s.keep, 0)
}
