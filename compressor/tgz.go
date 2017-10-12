package compressor

import (
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/helper"
)

// Tgz .tar.gz compressor
//
// type: tgz
type Tgz struct {
}

func (ctx *Tgz) perform(model config.ModelConfig) (archivePath string, err error) {
	filePath := archiveFilePath(model, ".tar.gz")
	_, err = helper.Exec("tar", "zcf", filePath, model.Name)
	if err == nil {
		archivePath = filePath
		return
	}
	return
}
