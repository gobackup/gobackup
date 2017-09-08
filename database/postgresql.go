package database

import (
	"fmt"
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/helper"
	"github.com/huacnlee/gobackup/logger"
	"os/exec"
	"path"
	"strings"
)

// PostgreSQL database
type PostgreSQL struct {
	Name        string
	host        string
	port        string
	database    string
	username    string
	password    string
	dumpCommand string
	dumpPath    string
	model       config.ModelConfig
}

func (ctx PostgreSQL) perform(model config.ModelConfig, dbCfg config.SubConfig) (err error) {
	viper := dbCfg.Viper
	viper.SetDefault("host", "localhost")
	viper.SetDefault("port", 5432)

	ctx.Name = dbCfg.Name
	ctx.host = viper.GetString("host")
	ctx.port = viper.GetString("port")
	ctx.database = viper.GetString("database")
	ctx.username = viper.GetString("username")
	ctx.password = viper.GetString("password")
	ctx.model = model

	if err = ctx.prepare(); err != nil {
		return
	}

	logger.Info("=> database | PostgreSQL:", ctx.Name)
	err = ctx.dump()
	return
}

func (ctx *PostgreSQL) prepare() (err error) {
	ctx.dumpPath = path.Join(ctx.model.DumpPath, "postgresql")
	helper.MkdirP(ctx.dumpPath)

	// mysqldump command
	dumpArgs := []string{}
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

	ctx.dumpCommand = "pg_dump " + strings.Join(dumpArgs, " ") + " " + ctx.database

	if len(ctx.password) > 0 {
		exec.Command("export", "PGPASSWORD="+ctx.password).Run()
	}

	return nil
}

func (ctx *PostgreSQL) dump() error {
	dumpFilePath := path.Join(ctx.dumpPath, ctx.database+".sql")
	logger.Info("-> Dumping PostgreSQL...")
	_, err := helper.Exec(ctx.dumpCommand, "-f", dumpFilePath)
	if err != nil {
		return err
	}
	logger.Info("dump path:", dumpFilePath)
	return nil
}
