package storage

import (
	"fmt"
	"github.com/huacnlee/gobackup/config"
	"path/filepath"
)

// Base storage
type Base interface {
	perform(model config.ModelConfig, fileKey, archivePath string) error
}

// Run storage
func Run(model config.ModelConfig, archivePath string) error {
	var ctx Base
	switch model.StoreWith.Type {
	case "local":
		ctx = &Local{}
	case "ftp":
		ctx = &FTP{}
	case "scp":
		ctx = &SCP{}
	case "s3":
		ctx = &S3{}
	case "oss":
		ctx = &OSS{}
	default:
		return fmt.Errorf("[%s] storage type has not implement", model.StoreWith.Type)
	}

	fileKey := filepath.Base(archivePath)
	err := ctx.perform(model, fileKey, archivePath)
	if err != nil {
		return err
	}

	return nil
}
