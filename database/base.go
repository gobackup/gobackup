package database

import (
	"fmt"
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/logger"
)

// Base interface
type Base interface {
	perform(model config.ModelConfig, dbConfig config.SubConfig) error
}

// New - initialize Database
func runModel(model config.ModelConfig, dbConfig config.SubConfig) (err error) {
	var ctx Base
	switch dbConfig.Type {
	case "mysql":
		ctx = &MySQL{}
	case "redis":
		ctx = &Redis{}
	case "postgresql":
		ctx = &PostgreSQL{}
	case "mongodb":
		ctx = &MongoDB{}
	default:
		logger.Warn(fmt.Errorf("model: %s databases.%s config `type: %s`, but is not implement", model.Name, dbConfig.Name, dbConfig.Type))
		return
	}

	// perform
	err = ctx.perform(model, dbConfig)
	if err != nil {
		return err
	}
	logger.Info("")

	return
}

// Run databases
func Run(model config.ModelConfig) error {
	if len(model.Databases) == 0 {
		return nil
	}

	logger.Info("------------- Databases -------------")
	for _, dbCfg := range model.Databases {
		err := runModel(model, dbCfg)
		if err != nil {
			return err
		}
	}
	logger.Info("------------- Databases -------------\n")

	return nil
}
