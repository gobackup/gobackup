package database

import (
	"fmt"
	"path"
	"time"

	"github.com/spf13/viper"

	"github.com/gobackup/gobackup/config"
	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/hook"
	"github.com/gobackup/gobackup/logger"
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
	// Init database config, prepare all things
	init() error
	// Peform backup
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

func runHook(action, script string) error {
	logger := logger.Tag("Database")
	if len(script) == 0 {
		return nil
	}
	logger.Infof("Run %s", action)

	if _, err := helper.ExecScriptWithStdio(script, true); err != nil {
		return fmt.Errorf("Run %s failed: %v", action, err)
	}

	logger.Infof("Run %s succeeded", action)
	return nil
}

// New - initialize Database
func runModel(model config.ModelConfig, dbConfig config.SubConfig) (err error) {
	logger := logger.Tag("Database")

	base := newBase(model, dbConfig)
	var db Database
	switch dbConfig.Type {
	case "mysql":
		db = &MySQL{Base: base}
	case "mariadb":
		db = &MariaDB{Base: base}
	case "redis":
		db = &Redis{Base: base}
	case "postgresql":
		db = &PostgreSQL{Base: base}
	case "mongodb":
		db = &MongoDB{Base: base}
	case "sqlite":
		db = &SQLite{Base: base}
	case "mssql":
		db = &MSSQL{Base: base}
	case "influxdb2":
		db = &InfluxDB2{Base: base}
	case "etcd":
		db = &Etcd{Base: base}
	case "firebird":
		db = &Firebird{Base: base}
	default:
		logger.Warn(fmt.Errorf("model: %s databases.%s config `type: %s`, but is not implement", model.Name, dbConfig.Name, dbConfig.Type))
		return
	}

	logger.Infof("=> database | %v: %v", dbConfig.Type, base.name)

	beforeScript := dbConfig.Viper.GetString("before_script")
	afterScript := dbConfig.Viper.GetString("after_script")
	onExit := dbConfig.Viper.GetString("on_exit")
	ctx := hook.Context{Model: model.Name, Stage: hook.StageDatabase, Step: dbConfig.Name}

	startTime := time.Now()

	// before perform: legacy before_script first, then before hooks
	beforeErr := runHook("dump before_script", beforeScript)
	if beforeErr == nil {
		beforeErr = hook.RunBefore(dbConfig.Hooks, ctx)
	}

	// perform
	err = beforeErr
	if err == nil {
		if err = db.init(); err == nil {
			err = db.perform()
		}
	}

	// legacy after_script (keeps its on_exit semantics, skipped when before
	// failed, as before). Its failure is the step's failure.
	if beforeErr == nil {
		err = runLegacyAfter(err, afterScript, onExit)
	}

	// after hooks see the final status of the step
	hook.RunAfter(dbConfig.Hooks, ctx, hook.Result{Err: err, Duration: time.Since(startTime)})

	return err
}

// runLegacyAfter runs legacy `after_script` according to `on_exit` and returns
// the error to report: the dump error, or the after_script error if it fails.
//
// Kept as-is for backward compatibility: when the dump succeeds after_script
// runs regardless of on_exit; when it fails after_script runs only with
// on_exit always / failure.
func runLegacyAfter(performErr error, afterScript, onExit string) error {
	logger := logger.Tag("Database")

	if performErr != nil {
		logger.Info("Dump failed")
		if len(afterScript) == 0 {
			return performErr
		} else if len(onExit) != 0 {
			switch onExit {
			case "always":
				logger.Info("on_exit is always, start to run after_script")
			case "success":
				logger.Info("on_exit is success, skip run after_script")
				return performErr
			case "failure":
				logger.Info("on_exit is failure, start to run after_script")
			default:
				// skip after
				return performErr
			}
		} else {
			return performErr
		}
	} else {
		logger.Info("Dump succeeded")
	}

	if err := runHook("dump after_script", afterScript); err != nil {
		return err
	}

	return performErr
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
