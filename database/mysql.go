package database

import (
	"fmt"
	"github.com/huacnlee/gobackup/helper"
	"github.com/huacnlee/gobackup/logger"
	"path"
)

// MySQL database
//
// type: mysql
// host: 127.0.0.1
// port: 3306
// database:
// username: root
// password:
// additionalOptions:
type MySQL struct {
	Base
	host              string
	port              string
	database          string
	username          string
	password          string
	additionalOptions string
}

func (ctx *MySQL) perform() (err error) {
	viper := ctx.viper
	viper.SetDefault("host", "127.0.0.1")
	viper.SetDefault("username", "root")
	viper.SetDefault("port", 3306)

	ctx.host = viper.GetString("host")
	ctx.port = viper.GetString("port")
	ctx.database = viper.GetString("database")
	ctx.username = viper.GetString("username")
	ctx.password = viper.GetString("password")
	ctx.additionalOptions = viper.GetString("additional_options")

	// mysqldump command
	if len(ctx.database) == 0 {
		return fmt.Errorf("mysql database config is required")
	}

	err = ctx.dump()
	return
}

func (ctx *MySQL) dumpArgs() []string {
	dumpArgs := []string{}
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
		dumpArgs = append(dumpArgs, `-p`+ctx.password)
	}
	if len(ctx.additionalOptions) > 0 {
		dumpArgs = append(dumpArgs, ctx.additionalOptions)
	}

	dumpArgs = append(dumpArgs, ctx.database)
	dumpFilePath := path.Join(ctx.dumpPath, ctx.database+".sql")
	dumpArgs = append(dumpArgs, "--result-file="+dumpFilePath)
	return dumpArgs
}

func (ctx *MySQL) dump() error {
	logger.Info("-> Dumping MySQL...")
	_, err := helper.Exec("mysqldump", ctx.dumpArgs()...)
	if err != nil {
		return fmt.Errorf("-> Dump error: %s", err)
	}
	logger.Info("dump path:", ctx.dumpPath)
	return nil
}
