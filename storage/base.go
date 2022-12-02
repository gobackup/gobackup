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
	cycler      *Cycler
}

// Storage interface
type Storage interface {
	open() error
	close()
	upload(fileKey string) error
	delete(fileKey string) error
}

func newBase(model config.ModelConfig, archivePath string, storageConfig config.SubConfig) (base Base) {
	// Backward compatible with `store_with` config
	var cyclerName string
	if storageConfig.Name == "" {
		cyclerName = model.Name
	} else {
		cyclerName = fmt.Sprintf("%s_%s", model.Name, storageConfig.Name)
	}

	base = Base{
		model:       model,
		archivePath: archivePath,
		viper:       storageConfig.Viper,
		cycler:      &Cycler{name: cyclerName},
	}

	if base.viper != nil {
		base.keep = base.viper.GetInt("keep")
	}

	return
}

// run storage
func runModel(model config.ModelConfig, archivePath string, storageConfig config.SubConfig) (err error) {
	logger := logger.Tag("Storage")

	newFileKey := filepath.Base(archivePath)
	base := newBase(model, archivePath, storageConfig)
	var s Storage
	switch storageConfig.Type {
	case "local":
		s = &Local{Base: base}
	case "ftp":
		s = &FTP{Base: base}
	case "scp":
		s = &SCP{Base: base}
	case "oss":
		s = &OSS{Base: base}
	case "gcs":
		s = &GCS{Base: base}
	case "s3":
		s = &S3{Base: base, Service: "s3"}
	case "b2":
		s = &S3{Base: base, Service: "b2"}
	case "us3":
		s = &S3{Base: base, Service: "us3"}
	case "cos":
		s = &S3{Base: base, Service: "cos"}
	case "kodo":
		s = &S3{Base: base, Service: "kodo"}
	case "r2":
		s = &S3{Base: base, Service: "r2"}
	case "spaces":
		s = &S3{Base: base, Service: "spaces"}
	default:
		return fmt.Errorf("[%s] storage type has not implement", storageConfig.Type)
	}

	logger.Info("=> Storage | " + storageConfig.Type)
	err = s.open()
	if err != nil {
		return err
	}
	defer s.close()

	err = s.upload(newFileKey)
	if err != nil {
		return err
	}

	base.cycler.run(newFileKey, base.keep, s.delete)

	return nil
}

// Run storage
func Run(model config.ModelConfig, archivePath string) (err error) {
	for _, storageConfig := range model.Storages {
		err := runModel(model, archivePath, storageConfig)
		if err != nil {
			return err
		}
	}

	return nil
}
