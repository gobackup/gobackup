package database

import (
	"fmt"
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/logger"
)

// Base interface
type Base interface {
	perform() error
}

// New - initialize Database
func runModel(subCfg config.SubConfig) (err error) {
	var ctx Base
	switch subCfg.Type {
	case "mysql":
		ctx = newMySQL(subCfg)
	case "redis":
		ctx = newRedis()
	default:
		logger.Warn(fmt.Errorf("databases.%s config `type: %s`, but is not implement", subCfg.Name, subCfg.Type))
		return
	}

	err = ctx.perform()

	return
}

// Run databases
func Run(cfg config.Config) (err error) {
	for _, dbCfg := range cfg.Databases {
		err = runModel(dbCfg)
		if err != nil {
			return
		}
	}

	return
}
