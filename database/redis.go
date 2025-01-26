package database

import (
	"fmt"
	"path"
	"regexp"
	"strings"

	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
)

type redisMode int

const (
	redisModeSync redisMode = iota
	redisModeCopy
)

// Redis database
//
// type: redis
// mode: sync # or copy for use rdb_path
// invoke_save: true
// host: 192.168.1.2
// port: 6379
// socket:
// password:
// rdb_path: /var/db/redis/dump.rdb
type Redis struct {
	Base
	host       string
	port       string
	socket     string
	password   string
	mode       redisMode
	invokeSave bool
	// path of rdb file, example: /var/lib/redis/dump.rdb
	rdbPath string
	args    string

	_dumpFilePath string
}

func (db *Redis) init() (err error) {

	viper := db.viper
	viper.SetDefault("rdb_path", "/var/db/redis/dump.rdb")
	viper.SetDefault("host", "127.0.0.1")
	viper.SetDefault("port", "6379")
	viper.SetDefault("invoke_save", true)
	viper.SetDefault("mode", "copy")

	db.host = viper.GetString("host")
	db.port = viper.GetString("port")
	db.socket = viper.GetString("socket")
	db.password = viper.GetString("password")
	db.rdbPath = viper.GetString("rdb_path")
	db.invokeSave = viper.GetBool("invoke_save")
	db.args = viper.GetString("args")

	// Force set invokeSave = false, when mode = copy
	if viper.GetString("mode") == "copy" {
		db.invokeSave = false
	}

	// socket
	if len(db.socket) != 0 {
		db.host = ""
		db.port = ""
	}

	if viper.GetString("mode") == "sync" {
		db.mode = redisModeSync
	} else {
		db.mode = redisModeCopy
	}

	db._dumpFilePath = path.Join(db.dumpPath, "dump.rdb")

	return nil
}

func (db *Redis) build() string {
	if db.mode == redisModeCopy {
		return strings.Join([]string{
			"cp",
			db.rdbPath,
			db._dumpFilePath,
		}, " ")
	}

	// redis-cli command
	args := []string{"redis-cli"}
	if len(db.host) > 0 {
		args = append(args, "-h "+db.host)
	}
	if len(db.port) > 0 {
		args = append(args, "-p "+db.port)
	}
	if len(db.socket) > 0 {
		args = append(args, "-s", db.socket)
	}
	if len(db.password) > 0 {
		args = append(args, `-a `+db.password)
	}

	if len(db.args) > 0 {
		args = append(args, db.args)
	}

	args = append(args, "--rdb", db._dumpFilePath)

	return strings.Join(args, " ")
}

func (db *Redis) perform() (err error) {
	if db.mode == redisModeCopy {
		if !helper.IsExistsPath(db.rdbPath) {
			return fmt.Errorf("Redis RDB file: %s does not exist", db.rdbPath)
		}
	}

	if err = db.trySave(); err != nil {
		return
	}

	if db.mode == redisModeCopy {
		err = db.copy()
	} else {
		err = db.sync()
	}

	return err
}

func (db *Redis) trySave() error {
	logger := logger.Tag("Redis")

	if !db.invokeSave {
		return nil
	}

	// FIXME: add retry
	logger.Info("Perform redis-cli save...")
	out, err := helper.Exec(db.build(), "SAVE")
	if err != nil {
		return fmt.Errorf("redis-cli SAVE failed %s", err)
	}

	if !regexp.MustCompile("OK$").MatchString(strings.TrimSpace(out)) {
		return fmt.Errorf(`failed to invoke the "SAVE" command Response was: %s`, out)
	}

	return nil
}

func (db *Redis) sync() error {
	logger := logger.Tag("Redis")

	logger.Info("Syncing redis dump to", db._dumpFilePath)
	_, err := helper.Exec(db.build())
	if err != nil {
		return fmt.Errorf("dump redis error: %s", err)
	}

	if !helper.IsExistsPath(db._dumpFilePath) {
		return fmt.Errorf("dump result file %s not found", db._dumpFilePath)
	}

	return nil
}

func (db *Redis) copy() error {
	logger := logger.Tag("Redis")

	logger.Info("Copying redis dump to", db._dumpFilePath)
	_, err := helper.Exec(db.build())
	if err != nil {
		return fmt.Errorf("copy redis dump file error: %s", err)
	}
	return nil
}
