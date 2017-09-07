package database

import (
	"fmt"
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/logger"
)

// Base interface
type Base interface {
	perform() error
	prepare() error
}

// New - initialize Database
func runModel(subCfg config.SubConfig) (err error) {
	var ctx Base
	switch subCfg.Type {
	case "mysql":
		ctx = newMySQL(subCfg)
	case "redis":
		ctx = newRedis(subCfg)
	default:
		logger.Warn(fmt.Errorf("databases.%s config `type: %s`, but is not implement", subCfg.Name, subCfg.Type))
		return
	}
	err = ctx.prepare()
	if err != nil {
		logger.Error(err)
	}
	err = ctx.perform()
	if err != nil {
		logger.Error(err)
	}
	logger.Info("")

	return
}

// Run databases
func Run() {
	logger.Info("------------- Databases --------------")
	for _, dbCfg := range config.Databases {
		runModel(dbCfg)
	}
	logger.Info("------------- End databases --------------")

	return
}
