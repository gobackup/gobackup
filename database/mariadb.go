package database

import (
	"fmt"
	"strings"
	"os"
	"path"

	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
)

// Mariadb database
//
// type: mariadb
// host: 127.0.0.1
// port: 3306
// socket:
// databases:
// username: root
// password:
// args:

type MariaDB struct {
	Base
	host          		string
	port          		string
	socket        		string
	databases     		[]string
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
	db.databases = viper.GetStringSlice("databases")
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

func createdatabasesfile(databasesfile string, databases []string) error {
	file, err := os.Create(databasesfile)
	if err != nil {
		return err
	}
	defer file.Close()
	for _, database := range databases {
		_, err := file.WriteString(database + "\n")
		if err != nil {
			return err
		}
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

	if len(db.databases) > 0 {
		databasesfile := path.Join(db.dumpPath, "databases-file.txt")
		err := createdatabasesfile(databasesfile, db.databases)
		if err != nil {
			logger.Error("-> Dump error: %s", err)
		}
		dumpArgs = append(dumpArgs, "--databases-file=" + databasesfile)
	}

	if len(db.args) > 0 {
		dumpArgs = append(dumpArgs, db.args)
	}

	dumpArgs = append(dumpArgs, "--target-dir="+db.dumpPath)
	return "mariadb-backup --backup" + " " + strings.Join(dumpArgs, " ")
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
