package storage

import (
	"fmt"
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/logger"
	"github.com/spf13/viper"
	"path/filepath"
)

// Base storage
type Base struct {
	model       config.ModelConfig
	fileKey     string
	archivePath string
	viper       *viper.Viper
}

// Context storage interface
type Context interface {
	perform() error
}

func newBase(model config.ModelConfig, archivePath string) (base Base) {
	base = Base{
		model:       model,
		archivePath: archivePath,
		viper:       model.StoreWith.Viper,
		fileKey:     filepath.Base(archivePath),
	}
	return
}

// Run storage
func Run(model config.ModelConfig, archivePath string) error {
	logger.Info("------------- Storage --------------")

	base := newBase(model, archivePath)
	var ctx Context
	switch model.StoreWith.Type {
	case "local":
		ctx = &Local{Base: base}
	case "ftp":
		ctx = &FTP{Base: base}
	case "scp":
		ctx = &SCP{Base: base}
	case "s3":
		ctx = &S3{Base: base}
	case "oss":
		ctx = &OSS{Base: base}
	default:
		return fmt.Errorf("[%s] storage type has not implement", model.StoreWith.Type)
	}

	logger.Info("=> Storage | " + model.StoreWith.Type)

	err := ctx.perform()
	if err != nil {
		return err
	}

	logger.Info("------------- Storage --------------\n")
	return nil
}
