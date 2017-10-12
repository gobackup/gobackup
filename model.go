package main

import (
	"github.com/huacnlee/gobackup/archive"
	"github.com/huacnlee/gobackup/compressor"
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/database"
	"github.com/huacnlee/gobackup/encryptor"
	"github.com/huacnlee/gobackup/helper"
	"github.com/huacnlee/gobackup/logger"
	"github.com/huacnlee/gobackup/storage"
)

// Model class
type Model struct {
	Config config.ModelConfig
}

// Perform model
func (ctx Model) perform() {
	logger.Info("======== " + ctx.Config.Name + " ========")
	logger.Info("WorkDir:", ctx.Config.DumpPath+"\n")
	// defer ctx.cleanup()

	logger.Info("------------- Databases -------------")
	err := database.Run(ctx.Config)
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Info("------------- Databases -------------\n")

	if ctx.Config.Archive != nil {
		logger.Info("------------- Archives -------------")
		err = archive.Run(ctx.Config)
		if err != nil {
			logger.Error(err)
			return
		}
		logger.Info("------------- Archives -------------\n")
	}

	archivePath, err := compressor.Run(ctx.Config)
	if err != nil {
		logger.Error(err)
		return
	}

	archivePath, err = encryptor.Run(archivePath, ctx.Config)
	if err != nil {
		logger.Error(err)
		return
	}

	logger.Info("------------- Storage --------------")
	err = storage.Run(ctx.Config, archivePath)
	if err != nil {
		logger.Error(err)
		return
	}
	logger.Info("------------- Storage --------------\n")

}

// Cleanup model temp files
func (ctx Model) cleanup() {
	logger.Info("Cleanup temp dir...\n")
	helper.Exec("rm", "-rf", ctx.Config.DumpPath)
	logger.Info("======= End " + ctx.Config.Name + " =======\n\n")
}
