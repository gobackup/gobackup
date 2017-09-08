package compressor

import (
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/helper"
	"github.com/huacnlee/gobackup/logger"
	"os"
	"path"
	"time"
)

// Tgz .tar.gz compressor
type Tgz struct {
}

func (ctx *Tgz) perform(model config.ModelConfig) (resultPath *string, err error) {
	logger.Info("=> Compress with Tgz...")
	archivePath := path.Join(os.TempDir(), "gobackup", time.Now().Format(time.RFC3339)+".tar.gz")
	os.Chdir(model.DumpPath)
	_, err = helper.Exec("tar", "zcf", archivePath, "./")
	if err == nil {
		logger.Info("->", archivePath)
		resultPath = &archivePath
		return
	}
	return
}
