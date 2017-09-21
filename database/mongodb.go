package database

import (
	"fmt"
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/helper"
	"github.com/huacnlee/gobackup/logger"
	"path"
	"strings"
)

// MongoDB database
//
// type: mongodb
// host: 127.0.0.1
// port: 27017
// database:
// username:
// password:
// authdb:
// oplog: false
type MongoDB struct {
	Name     string
	host     string
	port     string
	database string
	username string
	password string
	authdb   string
	oplog    bool
	dumpPath string
}

var (
	mongodumpCli = "mongodump"
)

func (ctx *MongoDB) perform(model config.ModelConfig, dbCfg config.SubConfig) (err error) {
	viper := dbCfg.Viper
	viper.SetDefault("oplog", false)
	viper.SetDefault("host", "127.0.0.1")
	viper.SetDefault("username", "root")
	viper.SetDefault("port", 27017)

	ctx.Name = dbCfg.Name

	ctx.dumpPath = path.Join(model.DumpPath, "mongodb", ctx.Name)
	ctx.host = viper.GetString("host")
	ctx.port = viper.GetString("port")
	ctx.database = viper.GetString("database")
	ctx.username = viper.GetString("username")
	ctx.password = viper.GetString("password")
	ctx.oplog = viper.GetBool("oplog")
	ctx.authdb = viper.GetString("authdb")

	err = ctx.dump()
	if err != nil {
		return err
	}
	return nil
}

func (ctx *MongoDB) mongodump() string {
	return mongodumpCli + " " +
		ctx.nameOption() + " " +
		ctx.credentialOptions() + " " +
		ctx.connectivityOptions() + " " +
		ctx.oplogOption() + " " +
		"--out=" + ctx.dumpPath
}

func (ctx *MongoDB) nameOption() string {
	return "--db=" + ctx.database
}

func (ctx *MongoDB) credentialOptions() string {
	opts := []string{}
	if len(ctx.username) > 0 {
		opts = append(opts, "--username="+ctx.username)
	}
	if len(ctx.password) > 0 {
		opts = append(opts, "--password="+ctx.password)
	}
	if len(ctx.authdb) > 0 {
		opts = append(opts, "--authenticationDatabase="+ctx.authdb)
	}
	return strings.Join(opts, " ")
}

func (ctx *MongoDB) connectivityOptions() string {
	opts := []string{}
	if len(ctx.host) > 0 {
		opts = append(opts, "--host="+ctx.host+"")
	}
	if len(ctx.port) > 0 {
		opts = append(opts, "--port="+ctx.port+"")
	}

	return strings.Join(opts, " ")
}

func (ctx *MongoDB) oplogOption() string {
	if ctx.oplog {
		return "--oplog"
	}

	return ""
}

func (ctx *MongoDB) dump() error {
	logger.Info("-> Dumping MongoDB...")
	helper.MkdirP(ctx.dumpPath)
	out, err := helper.Exec(ctx.mongodump())
	if err != nil {
		return fmt.Errorf("-> Dump error: %s", err)
	}
	logger.Info(out)
	logger.Info("dump path:", ctx.dumpPath)
	return nil
}
