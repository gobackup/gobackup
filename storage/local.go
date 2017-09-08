package storage

import (
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/helper"
	"github.com/huacnlee/gobackup/logger"
)

// Local storage
type Local struct {
}

func (ctx *Local) perform(archivePath string) error {
	logger.Info("=> storage | Local")
	destPath := config.StoreWith.Viper.GetString("path")
	helper.MkdirP(destPath)
	_, err := helper.Exec("cp", archivePath, destPath)
	if err != nil {
		return err
	}
	logger.Info("Store successed", destPath)
	return nil
}
