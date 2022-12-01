package database

import (
	"fmt"
	"path"
	"strings"

	"github.com/huacnlee/gobackup/helper"
	"github.com/huacnlee/gobackup/logger"
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
// additional_options:
type MySQL struct {
	Base
	host              string
	port              string
	socket            string
	database          string
	username          string
	password          string
	additionalOptions []string
}

func (db *MySQL) perform() (err error) {
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
	addOpts := viper.GetString("additional_options")
	if len(addOpts) > 0 {
		db.additionalOptions = strings.Split(addOpts, " ")
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

	err = db.dump()
	return
}

func (db *MySQL) dumpArgs() []string {
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
	if len(db.additionalOptions) > 0 {
		dumpArgs = append(dumpArgs, db.additionalOptions...)
	}

	dumpArgs = append(dumpArgs, db.database)
	dumpFilePath := path.Join(db.dumpPath, db.database+".sql")
	dumpArgs = append(dumpArgs, "--result-file="+dumpFilePath)
	return dumpArgs
}

func (db *MySQL) dump() error {
	logger := logger.Tag("MySQL")

	logger.Info("-> Dumping MySQL...")
	_, err := helper.Exec("mysqldump", db.dumpArgs()...)
	if err != nil {
		return fmt.Errorf("-> Dump error: %s", err)
	}
	logger.Info("dump path:", db.dumpPath)
	return nil
}
