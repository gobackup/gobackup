package main

import (
	"github.com/huacnlee/gobackup/compressor"
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/database"
	"github.com/huacnlee/gobackup/helper"
	"github.com/huacnlee/gobackup/logger"
)

func main() {
	defer cleanup()
	logger.Info("WorkDir:", config.DumpPath)
	err := database.Run()
	if err != nil {
		logger.Error(err)
		return
	}

	err = compressor.Run()
	if err != nil {
		logger.Error(err)
		return
	}
}

func cleanup() {
	logger.Info("Cleanup temp dir...")
	helper.Exec("rm", "-rf", config.DumpPath)
}
