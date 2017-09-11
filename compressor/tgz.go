package compressor

import (
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/helper"
	"github.com/huacnlee/gobackup/logger"
)

// Tgz .tar.gz compressor
type Tgz struct {
}

func (ctx *Tgz) perform(model config.ModelConfig) (archivePath string, err error) {
	logger.Info("=> Compress with Tgz...")
	filePath := archiveFilePath(".tar.gz")
	_, err = helper.Exec("tar", "zcf", filePath, model.Name)
	if err == nil {
		archivePath = filePath
		return
	}
	return
}
