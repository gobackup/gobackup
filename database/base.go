package database

import (
	"fmt"
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/helper"
	"github.com/huacnlee/gobackup/logger"
	"github.com/spf13/viper"
	"path"
)

// Base database
type Base struct {
	model    config.ModelConfig
	dbConfig config.SubConfig
	viper    *viper.Viper
	name     string
	dumpPath string
}

// Context database interface
type Context interface {
	perform() error
}

func newBase(model config.ModelConfig, dbConfig config.SubConfig) (base Base) {
	base = Base{
		model:    model,
		dbConfig: dbConfig,
		viper:    dbConfig.Viper,
		name:     dbConfig.Name,
	}
	base.dumpPath = path.Join(model.DumpPath, dbConfig.Type, base.name)
	helper.MkdirP(base.dumpPath)
	return
}

// New - initialize Database
func runModel(model config.ModelConfig, dbConfig config.SubConfig) (err error) {
	base := newBase(model, dbConfig)
	var ctx Context
	switch dbConfig.Type {
	case "mysql":
		ctx = &MySQL{Base: base}
	case "redis":
		ctx = &Redis{Base: base}
	case "postgresql":
		ctx = &PostgreSQL{Base: base}
	case "mongodb":
		ctx = &MongoDB{Base: base}
	default:
		logger.Warn(fmt.Errorf("model: %s databases.%s config `type: %s`, but is not implement", model.Name, dbConfig.Name, dbConfig.Type))
		return
	}

	logger.Info("=> database |", dbConfig.Type, ":", base.name)

	// perform
	err = ctx.perform()
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
