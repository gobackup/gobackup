package database

import (
	"fmt"
	"github.com/huacnlee/gobackup/helper"
	"github.com/huacnlee/gobackup/logger"
	"os"
	"path"
	"strings"
)

// PostgreSQL database
//
// type: postgresql
// host: localhost
// port: 5432
// database: test
// username:
// password:
type PostgreSQL struct {
	Base
	host        string
	port        string
	database    string
	username    string
	password    string
	dumpCommand string
	params      string
}

func (ctx PostgreSQL) perform() (err error) {
	viper := ctx.viper
	viper.SetDefault("host", "localhost")
	viper.SetDefault("port", 5432)
	viper.SetDefault("params", "")

	ctx.host = viper.GetString("host")
	ctx.port = viper.GetString("port")
	ctx.database = viper.GetString("database")
	ctx.username = viper.GetString("username")
	ctx.password = viper.GetString("password")
	ctx.params = viper.GetString("params")

	if err = ctx.prepare(); err != nil {
		return
	}

	err = ctx.dump()
	return
}

func (ctx *PostgreSQL) prepare() (err error) {
	// pg_dump command
	var dumpArgs []string
	if len(ctx.database) == 0 {
		return fmt.Errorf("PostgreSQL database config is required")
	}
	if len(ctx.host) > 0 {
		dumpArgs = append(dumpArgs, "--host="+ctx.host)
	}
	if len(ctx.port) > 0 {
		dumpArgs = append(dumpArgs, "--port="+ctx.port)
	}
	if len(ctx.username) > 0 {
		dumpArgs = append(dumpArgs, "--username="+ctx.username)
	}
	if len(ctx.params) > 0 {
		dumpArgs = append(dumpArgs, ctx.params)
	}

	ctx.dumpCommand = "pg_dump " + strings.Join(dumpArgs, " ") + " " + ctx.database

	return nil
}

func (ctx *PostgreSQL) dump() error {
	dumpFilePath := path.Join(ctx.dumpPath, ctx.database+".sql")
	logger.Info("-> Dumping PostgreSQL...")
	if len(ctx.password) > 0 {
		os.Setenv("PGPASSWORD", ctx.password)
	}
	_, err := helper.Exec(ctx.dumpCommand, "-f", dumpFilePath)
	if err != nil {
		return err
	}
	logger.Info("dump path:", dumpFilePath)
	return nil
}
