package database

import (
	"fmt"
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/helper"
	"github.com/huacnlee/gobackup/logger"
	"path"
	"regexp"
	"strings"
)

type redisMode int

const (
	redisModeSync redisMode = iota
	redisModeCopy
)

// Redis database
type Redis struct {
	Name       string
	host       string
	port       string
	password   string
	mode       redisMode
	invokeSave bool
	// path of rdb file, example: /var/lib/redis/dump.rdb
	rdbPath  string
	dumpPath string
}

var (
	redisCliCommand = "redis-cli"
)

func newRedis(dbCfg config.SubConfig) (ctx *Redis) {
	viper := dbCfg.Viper
	viper.SetDefault("rdb_path", "/var/db/redis/dump.rdb")
	viper.SetDefault("host", "127.0.0.1")
	viper.SetDefault("port", "6379")
	viper.SetDefault("invoke_save", true)
	viper.SetDefault("mode", "copy")

	ctx = &Redis{
		Name:       dbCfg.Name,
		host:       viper.GetString("host"),
		port:       viper.GetString("port"),
		password:   viper.GetString("password"),
		rdbPath:    viper.GetString("rdb_path"),
		invokeSave: viper.GetBool("invoke_save"),
	}

	if viper.GetString("mode") == "sync" {
		ctx.mode = redisModeSync
	} else {
		ctx.mode = redisModeCopy
	}

	ctx.dumpPath = path.Join(config.DumpPath, "databases", "redis")
	return
}

// Perform redis
func (ctx *Redis) perform() error {
	logger.Info("=> database | Redis:", ctx.Name)
	if !helper.IsExistsPath(ctx.rdbPath) {
		return fmt.Errorf("Redis RDB file: %s does not exist", ctx.rdbPath)
	}

	err := ctx.save()
	if err != nil {
		return err
	}

	if ctx.mode == redisModeCopy {
		err = ctx.copy()
	} else {
		err = ctx.sync()
	}
	if err != nil {
		return err
	}
	return nil
}

func (ctx *Redis) prepare() error {
	helper.MkdirP(ctx.dumpPath)

	// redis-cli command
	args := []string{"redis-cli"}
	if len(ctx.host) > 0 {
		args = append(args, "-h "+ctx.host)
	}
	if len(ctx.port) > 0 {
		args = append(args, "-p "+ctx.port)
	}
	if len(ctx.password) > 0 {
		args = append(args, "-a "+ctx.password)
	}
	redisCliCommand = strings.Join(args, " ")

	return nil
}

func (ctx *Redis) save() error {
	if !ctx.invokeSave {
		return nil
	}
	// FIXME: add retry
	logger.Info("Perform redis-cli save...")
	out, err := helper.Exec(redisCliCommand, "SAVE")
	if err != nil {
		return fmt.Errorf("redis-cli SAVE failed %s", err)
	}

	if !regexp.MustCompile("OK$").MatchString(strings.TrimSpace(out)) {
		return fmt.Errorf(`Failed to invoke the "SAVE" command Response was: %s`, out)
	}

	return nil
}

func (ctx *Redis) sync() error {
	dumpFilePath := path.Join(ctx.dumpPath, "dump.rdb")
	logger.Info("Syncing redis dump to", dumpFilePath)
	_, err := helper.Exec(redisCliCommand, "--rdb", "-", "|", "cat >", dumpFilePath)
	if err != nil {
		return fmt.Errorf("dump redis error: %s", err)
	}
	return nil
}

func (ctx *Redis) copy() error {
	logger.Info("Copying redis dump to", ctx.dumpPath)
	_, err := helper.Exec("cp", ctx.rdbPath, ctx.dumpPath)
	if err != nil {
		return fmt.Errorf("copy redis dump file error: %s", err)
	}
	return nil
}
