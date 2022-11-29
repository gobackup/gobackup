package storage

import (
	"fmt"
	"path/filepath"

	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/logger"
	"github.com/spf13/viper"
)

// Base storage
type Base struct {
	model       config.ModelConfig
	archivePath string
	viper       *viper.Viper
	keep        int
}

// Context storage interface
type Service interface {
	open() error
	close()
	upload(fileKey string) error
	delete(fileKey string) error
}

func newBase(model config.ModelConfig, archivePath string) (base Base) {
	base = Base{
		model:       model,
		archivePath: archivePath,
		viper:       model.StoreWith.Viper,
	}

	if base.viper != nil {
		base.keep = base.viper.GetInt("keep")
	}

	return
}

// Run storage
func Run(model config.ModelConfig, archivePath string) (err error) {
	logger := logger.Tag("Storage")

	newFileKey := filepath.Base(archivePath)
	base := newBase(model, archivePath)
	var s Service
	switch model.StoreWith.Type {
	case "local":
		s = &Local{Base: base}
	case "ftp":
		s = &FTP{Base: base}
	case "scp":
		s = &SCP{Base: base}
	case "s3":
		s = &S3{Base: base, Provider: "S3"}
	case "b2":
		s = &S3{Base: base, Provider: "B2"}
	case "oss":
		s = &OSS{Base: base}
	case "gcs":
		s = &GCS{Base: base}
	default:
		return fmt.Errorf("[%s] storage type has not implement", model.StoreWith.Type)
	}

	logger.Info("=> Storage | " + model.StoreWith.Type)
	err = s.open()
	if err != nil {
		return err
	}
	defer s.close()

	err = s.upload(newFileKey)
	if err != nil {
		return err
	}

	cycler := Cycler{}
	cycler.run(model.Name, newFileKey, base.keep, s.delete)

	return nil
}
