package storage

import (
	"github.com/huacnlee/gobackup/helper"
	"github.com/huacnlee/gobackup/logger"
)

// Local storage
//
// type: local
// path: /data/backups
type Local struct {
	Base
}

func (ctx *Local) perform() error {
	destPath := ctx.model.StoreWith.Viper.GetString("path")
	helper.MkdirP(destPath)
	_, err := helper.Exec("cp", ctx.archivePath, destPath)
	if err != nil {
		return err
	}
	logger.Info("Store successed", destPath)
	return nil
}
