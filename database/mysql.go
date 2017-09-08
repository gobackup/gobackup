package database

import (
	"fmt"
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/helper"
	"github.com/huacnlee/gobackup/logger"
	"path"
	"strings"
)

// MySQL database
type MySQL struct {
	Name        string
	host        string
	port        string
	database    string
	username    string
	password    string
	dumpCommand string
	dumpPath    string
}

// NewMySQL instrance
func newMySQL(dbCfg config.SubConfig) (ctx *MySQL) {
	viper := dbCfg.Viper
	viper.SetDefault("host", "localhost")
	viper.SetDefault("username", "root")
	viper.SetDefault("port", 3306)

	ctx = &MySQL{
		Name:     dbCfg.Name,
		host:     viper.GetString("host"),
		port:     viper.GetString("port"),
		database: viper.GetString("database"),
		username: viper.GetString("username"),
		password: viper.GetString("password"),
	}

	return ctx
}

func (ctx MySQL) perform() (err error) {
	logger.Info("=> database | MySQL:", ctx.Name)
	err = ctx.dump()
	return
}

func (ctx *MySQL) prepare() (err error) {
	ctx.dumpPath = path.Join(config.DumpPath, "mysql")
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

	ctx.dumpCommand = "mysqldump " + strings.Join(dumpArgs, " ") + " " + ctx.database

	return nil
}

func (ctx *MySQL) dump() error {
	dumpFilePath := path.Join(ctx.dumpPath, ctx.database+".sql")
	logger.Info("-> Dumping MySQL...")
	_, err := helper.Exec(ctx.dumpCommand, "--result-file="+dumpFilePath)
	if err != nil {
		return fmt.Errorf("-> Dump error: %s %s", ctx.dumpCommand, err)
	}
	logger.Info("dump path:", dumpFilePath)
	return nil
}
