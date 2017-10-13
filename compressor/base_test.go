package compressor

import (
	"github.com/huacnlee/gobackup/config"
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"strings"
	"testing"
	"time"
)

type Monkey struct {
}

func (ctx Monkey) perform(model config.ModelConfig) (archivePath string, err error) {
	result := "aaa"
	return result, nil
}

func TestArchiveFilePath(t *testing.T) {
	prefixPath := path.Join(os.TempDir(), "gobackup", time.Now().Format("2006.01.02.15.04"))
	assert.True(t, strings.HasPrefix(archiveFilePath(".tar"), prefixPath))
	assert.True(t, strings.HasSuffix(archiveFilePath(".tar"), ".tar"))
}

func TestBaseInterface(t *testing.T) {
	var ctx Base
	ctx = Monkey{}
	model := config.ModelConfig{
		Name: "TestMoneky",
	}
	result, err := ctx.perform(model)
	assert.Equal(t, result, "aaa")
	assert.Nil(t, err)
}
