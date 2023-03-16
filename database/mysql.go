package database

import (
	"fmt"
	"path"
	"strings"

	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
)

// MySQL database
//
// type: mysql
// host: 127.0.0.1
// port: 3306
// socket:
// database:
// username: root
// password:
// args:
type MySQL struct {
	Base
	host          string
	port          string
	socket        string
	database      string
	username      string
	password      string
	tables        []string
	excludeTables []string
	args          string
}

func (db *MySQL) init() (err error) {
	viper := db.viper
	viper.SetDefault("host", "127.0.0.1")
	viper.SetDefault("username", "root")
	viper.SetDefault("port", 3306)

	db.host = viper.GetString("host")
	db.port = viper.GetString("port")
	db.socket = viper.GetString("socket")
	db.database = viper.GetString("database")
	db.username = viper.GetString("username")
	db.password = viper.GetString("password")

	db.tables = viper.GetStringSlice("tables")
	db.excludeTables = viper.GetStringSlice("exclude_tables")

	if len(viper.GetString("args")) > 0 {
		db.args = viper.GetString("args")
	}

	// mysqldump command
	if len(db.database) == 0 {
		return fmt.Errorf("mysql database config is required")
	}

	// socket
	if len(db.socket) != 0 {
		db.host = ""
		db.port = ""
	}

	return nil
}

func (db *MySQL) build() string {
	dumpArgs := []string{}
	if len(db.host) > 0 {
		dumpArgs = append(dumpArgs, "--host", db.host)
	}
	if len(db.port) > 0 {
		dumpArgs = append(dumpArgs, "--port", db.port)
	}
	if len(db.socket) > 0 {
		dumpArgs = append(dumpArgs, "--socket", db.socket)
	}
	if len(db.username) > 0 {
		dumpArgs = append(dumpArgs, "-u", db.username)
	}
	if len(db.password) > 0 {
		dumpArgs = append(dumpArgs, `-p`+db.password)
	}

	for _, table := range db.excludeTables {
		dumpArgs = append(dumpArgs, "--ignore-table="+db.database+"."+table)
	}

	if len(db.args) > 0 {
		dumpArgs = append(dumpArgs, db.args)
	}

	dumpArgs = append(dumpArgs, db.database)
	if len(db.tables) > 0 {
		dumpArgs = append(dumpArgs, db.tables...)
	}

	dumpFilePath := path.Join(db.dumpPath, db.database+".sql")
	dumpArgs = append(dumpArgs, "--result-file="+dumpFilePath)

	return "mysqldump" + " " + strings.Join(dumpArgs, " ")
}

func (db *MySQL) perform() error {
	logger := logger.Tag("MySQL")

	logger.Info("-> Dumping MySQL...")
	_, err := helper.Exec(db.build())
	if err != nil {
		return fmt.Errorf("-> Dump error: %s", err)
	}
	logger.Info("dump path:", db.dumpPath)
	return nil
}
