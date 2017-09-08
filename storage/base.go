package storage

import (
	"github.com/huacnlee/gobackup/config"
)

// Base storage
type Base interface {
	perform(archivePath string) error
}

// Run storage
func Run(archivePath string) error {
	var ctx Base
	switch config.StoreWith.Type {
	case "local":
		ctx = &Local{}
	case "ftp":
		ctx = &FTP{}
	default:
		ctx = &Local{}
	}

	err := ctx.perform(archivePath)
	if err != nil {
		return err
	}

	return nil
}
