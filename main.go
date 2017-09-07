package main

import (
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/database"
	// "github.com/huacnlee/gobackup/logger"
)

var (
	appConfig config.Config
)

func main() {
	appConfig = config.LoadConfig()
	database.Run(appConfig)
}
