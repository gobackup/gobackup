package storage

import (
	"github.com/huacnlee/gobackup/config"
)

// Base storage
type Base interface {
	perform(model config.ModelConfig, archivePath string) error
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
	default:
		ctx = &Local{}
	}

	err := ctx.perform(model, archivePath)
	if err != nil {
		return err
	}

	return nil
}
