package database

import (
	"fmt"
	"path"

	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/helper"
	"github.com/huacnlee/gobackup/logger"
	"github.com/spf13/viper"
)

// Base database
type Base struct {
	model    config.ModelConfig
	dbConfig config.SubConfig
	viper    *viper.Viper
	name     string
	dumpPath string
}

// Database interface
type Database interface {
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
	if err := helper.MkdirP(base.dumpPath); err != nil {
		logger.Errorf("Failed to mkdir dump path %s: %v", base.dumpPath, err)
		return
	}
	return
}

// New - initialize Database
func runModel(model config.ModelConfig, dbConfig config.SubConfig) (err error) {
	logger := logger.Tag("Database")

	base := newBase(model, dbConfig)
	var db Database
	switch dbConfig.Type {
	case "mysql":
		db = &MySQL{Base: base}
	case "redis":
		db = &Redis{Base: base}
	case "postgresql":
		db = &PostgreSQL{Base: base}
	case "mongodb":
		db = &MongoDB{Base: base}
	default:
		logger.Warn(fmt.Errorf("model: %s databases.%s config `type: %s`, but is not implement", model.Name, dbConfig.Name, dbConfig.Type))
		return
	}

	logger.Infof("=> database | %v: %v", dbConfig.Type, base.name)

	// perform
	err = db.perform()
	if err != nil {
		return err
	}
	logger.Info("Dump succeeded")

	return
}

// Run databases
func Run(model config.ModelConfig) error {
	if len(model.Databases) == 0 {
		return nil
	}

	for _, dbCfg := range model.Databases {
		err := runModel(model, dbCfg)
		if err != nil {
			return err
		}
	}

	return nil
}
