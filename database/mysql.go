package database

import (
	"fmt"
	"path"
	"strings"

	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/helper"
	"github.com/huacnlee/gobackup/logger"
)

// MySQL database
//
// type: mysql
// host: localhost
// port: 3306
// database: test
// username: root
// password:
type MySQL struct {
	Name              string
	host              string
	port              string
	database          string
	username          string
	password          string
	dumpCommand       string
	dumpPath          string
	additionalOptions string
	model             config.ModelConfig
}

func (ctx *MySQL) perform(model config.ModelConfig, dbCfg config.SubConfig) (err error) {
	viper := dbCfg.Viper
	viper.SetDefault("host", "localhost")
	viper.SetDefault("username", "root")
	viper.SetDefault("port", 3306)
	viper.SetDefault("additional_options", "")

	ctx.Name = dbCfg.Name
	ctx.host = viper.GetString("host")
	ctx.port = viper.GetString("port")
	ctx.database = viper.GetString("database")
	ctx.username = viper.GetString("username")
	ctx.password = viper.GetString("password")
	ctx.additionalOptions = viper.GetString("additional_options")
	ctx.model = model

	if err = ctx.prepare(); err != nil {
		return
	}

	logger.Info("=> database | MySQL:", ctx.Name)
	err = ctx.dump()
	return
}

func (ctx *MySQL) prepare() (err error) {
	ctx.dumpPath = path.Join(ctx.model.DumpPath, "mysql", ctx.Name)
	helper.MkdirP(ctx.dumpPath)

	// mysqldump command
	dumpArgs := []string{}
	if len(ctx.database) == 0 {
		return fmt.Errorf("mysql database config is required")
	}
	if len(ctx.host) > 0 {
		dumpArgs = append(dumpArgs, "--host", ctx.host)
	}
	if len(ctx.port) > 0 {
		dumpArgs = append(dumpArgs, "--port", ctx.port)
	}
	if len(ctx.username) > 0 {
		dumpArgs = append(dumpArgs, "-u", ctx.username)
	}
	if len(ctx.password) > 0 {
		dumpArgs = append(dumpArgs, "-p"+ctx.password)
	}
	if len(ctx.additionalOptions) > 0 {
		dumpArgs = append(dumpArgs, ctx.additionalOptions)
	}

	dumpArgs = append(dumpArgs, ctx.database)

	ctx.dumpCommand = "mysqldump" + " " + strings.Join(dumpArgs, " ")

	return nil
}

func (ctx *MySQL) dump() error {
	dumpFilePath := path.Join(ctx.dumpPath, ctx.database+".sql")
	logger.Info("-> Dumping MySQL...")
	_, err := helper.Exec(ctx.dumpCommand, "--result-file="+dumpFilePath)
	if err != nil {
		return fmt.Errorf("-> Dump error: %s", err)
	}
	logger.Info("dump path:", dumpFilePath)
	return nil
}
