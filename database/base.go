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
	// prepare
	err = ctx.prepare()
	if err != nil {
		return err
	}

	// perform
	err = ctx.perform()
	if err != nil {
		return err
	}
	logger.Info("")

	return
}

// Run databases
func Run() error {
	logger.Info("------------- Databases --------------")
	for _, dbCfg := range config.Databases {
		err := runModel(dbCfg)
		if err != nil {
			return err
		}
	}
	logger.Info("----------- End databases ------------\n")

	return nil
}
