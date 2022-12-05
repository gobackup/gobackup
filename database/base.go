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

	// before perform
	if beforeScript := dbConfig.Viper.GetString("before_script"); len(beforeScript) != 0 {
		logger.Info("Run dump before_script")
		c, err := shlex.Split(beforeScript)
		if err != nil {
			return err
		}
		if _, err := helper.Exec(c[0], c[1:]...); err != nil {
			return err
		}
		logger.Info("Dump before_script succeeded")
	}

	afterScript := dbConfig.Viper.GetString("after_script")
	onExit := dbConfig.Viper.GetString("on_exit")

	// perform
	err = db.perform()
	if err != nil {
		logger.Info("Dump failed")
		if len(afterScript) == 0 {
			return
		} else if len(onExit) != 0 {
			switch onExit {
			case "always":
				logger.Info("on_exit is always, start to run after_script")
			case "success":
				logger.Info("on_exit is success, skip run after_script")
				return
			case "failure":
				logger.Info("on_exit is failure, start to run after_script")
			default:
				// skip after
				return
			}
		} else {
			return
		}
	} else {
		logger.Info("Dump succeeded")
	}

	// after perform
	if len(afterScript) != 0 {
		logger.Info("Run dump after_script")
		c, err := shlex.Split(afterScript)
		if err != nil {
			return err
		}
		if _, err := helper.Exec(c[0], c[1:]...); err != nil {
			return err
		}
		logger.Info("Dump after_script succeeded")
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
