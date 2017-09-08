package storage

import (
	"github.com/huacnlee/gobackup/config"
)

type Base interface {
	perform(archivePath string) error
}

func Run(archivePath string) error {
	var ctx Base
	switch config.StoreWith.Type {
	case "local":
		ctx = &Local{}
	default:
		ctx = &Local{}
	}

	err := ctx.perform(archivePath)
	if err != nil {
		return err
	}

	return nil
}
