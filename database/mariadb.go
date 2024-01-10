package database

import (
	"fmt"
	"strings"

	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
)

// Mariadb database
//
// type: mariadb
// host: 127.0.0.1
// port: 3306
// socket:
// database:
// username: root
// password:
// args:

type MariaDB struct {
	Base
	host          		string
	port          		string
	socket        		string
	database     		string
	username      		string
	password      		string
	args          		string
}

func (db *MariaDB) init() (err error) {
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

	if len(viper.GetString("args")) > 0 {
		db.args = viper.GetString("args")
	}

	// socket
	if len(db.socket) != 0 {
		db.host = ""
		db.port = ""
	}

	return nil
}

func (db *MariaDB) build() string {
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

	if len(db.args) > 0 {
		dumpArgs = append(dumpArgs, db.args)
	}
	if len(db.database) > 0 {
		dumpArgs = append(dumpArgs, "--databases="+db.database)
	}
	dumpArgs = append(dumpArgs, "--target-dir="+db.dumpPath)

	return "mariadb-backup --backup " + strings.Join(dumpArgs, " ")
}

func (db *MariaDB) perform() error {
	logger := logger.Tag("MariaDB")

	logger.Info("-> Dumping MariaDB...")
	_, err := helper.Exec(db.build())
	if err != nil {
		return fmt.Errorf("-> Dump error: %s", err)
	}
	logger.Info("dump path:", db.dumpPath)
	return nil
}
