package database

import (
	"fmt"
	"path"

	"github.com/google/shlex"
	"github.com/spf13/viper"

	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/helper"
	"github.com/huacnlee/gobackup/logger"
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

	// pre perform
	if preStart := dbConfig.Viper.GetString("prestart"); len(preStart) != 0 {
		logger.Info("Run dump prestart command")
		c, err := shlex.Split(preStart)
		if err != nil {
			return err
		}
		if _, err := helper.Exec(c[0], c[1:]...); err != nil {
			return err
		}
		logger.Info("Dump prestart command succeeded")
	}

	postStart := dbConfig.Viper.GetString("poststart")
	alwaysPostStart := dbConfig.Viper.GetBool("always_poststart")

	// perform
	err = db.perform()
	if err != nil {
		logger.Info("Dump failed")
		if len(postStart) == 0 {
			return
		} else if alwaysPostStart {
			logger.Info("always_poststart is true, start to run post start command")
		} else {
			return
		}
	} else {
		logger.Info("Dump succeeded")
	}

	// post perform
	if len(postStart) != 0 {
		logger.Info("Run dump poststart command")
		c, err := shlex.Split(postStart)
		if err != nil {
			return err
		}
		if _, err := helper.Exec(c[0], c[1:]...); err != nil {
			return err
		}
		logger.Info("Dump poststart command succeeded")
	}

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
