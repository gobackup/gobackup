package database

import (
	"fmt"
	"github.com/huacnlee/gobackup/logger"
	"os"
	"path"
	"regexp"
	"strings"
)

// Redis database
type Redis struct {
	redisCliPath string
	database     string
	// path of rdb file, example: /var/lib/redis/dump.rdb
	RdbFilePath string
}

var (
	redisCliCommand = "redis-cli"
	redisDumpPath   = path.Join(os.TempDir(), "databases", "redis")
)

func newRedis() (ctx *Redis) {
	ctx = &Redis{
		RdbFilePath: "/usr/local/var/db/redis/dump.rdb",
	}
	ctx.prepare()
	return
}

// Perform redis
func (ctx *Redis) Perform() error {
	logger.Info("Perform database/Redis")
	logger.Info("Redis dump path", redisDumpPath)
	if !isExistsPath(ctx.RdbFilePath) {
		return fmt.Errorf("Redis RDB file: %s does not exist", ctx.RdbFilePath)
	}

	err := ctx.save()
	if err != nil {
		return err
	}

	err = ctx.copy()
	if err != nil {
		return err
	}
	return nil
}

func (ctx *Redis) prepare() {
	ensureDir(redisDumpPath)
}

func (ctx *Redis) save() error {
	// FIXME: add retry
	logger.Info("Perform redis-cli save...")
	out, err := run(redisCliCommand, "SAVE")
	if err != nil {
		return fmt.Errorf("redis-cli SAVE failed %s", err)
	}

	if !regexp.MustCompile("OK$").MatchString(strings.TrimSpace(out)) {
		return fmt.Errorf(`Failed to invoke the "SAVE" command Response was: %s`, out)
	}

	return nil
}

func (ctx *Redis) copy() error {
	logger.Info("Copying redis dump to", redisDumpPath)
	_, err := run("cp", ctx.RdbFilePath, redisDumpPath)
	if err != nil {
		return fmt.Errorf("copy redis dump file error: %s", err)
	}
	return nil
}
