package compressor

import (
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/logger"
	"os"
	"path"
	"time"
)

// Base compressor
type Base interface {
	perform(model config.ModelConfig) (archivePath *string, err error)
}

func archiveFilePath(ext string) string {
	return path.Join(os.TempDir(), "gobackup", time.Now().Format(time.RFC3339)+ext)
}

// Run compressor
func Run(model config.ModelConfig) (archivePath *string, err error) {
	logger.Info("------------- Compressor --------------")
	var ctx Base
	switch model.CompressWith.Type {
	case "tgz":
		ctx = &Tgz{}
	default:
		ctx = &Tgz{}
	}

	// set workdir
	os.Chdir(path.Join(model.DumpPath, "../"))
	archivePath, err = ctx.perform(model)
	if err != nil {
		return
	}
	logger.Info("->", archivePath)
	logger.Info("----------- End Compressor ------------\n")

	return
}
