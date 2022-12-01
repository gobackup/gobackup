package database

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/huacnlee/gobackup/helper"
	"github.com/huacnlee/gobackup/logger"
)

// PostgreSQL database
//
// type: postgresql
// host: localhost
// port: 5432
// socket:
// database: test
// username:
// password:
type PostgreSQL struct {
	Base
	host        string
	port        string
	socket      string
	database    string
	username    string
	password    string
	dumpCommand string
	args        string
}

func (db PostgreSQL) perform() (err error) {
	viper := db.viper
	viper.SetDefault("host", "localhost")
	viper.SetDefault("port", 5432)
	viper.SetDefault("params", "")

	db.host = viper.GetString("host")
	db.port = viper.GetString("port")
	db.socket = viper.GetString("socket")
	db.database = viper.GetString("database")
	db.username = viper.GetString("username")
	db.password = viper.GetString("password")
	db.args = viper.GetString("args")

	// socket
	if len(db.socket) != 0 {
		db.host = ""
		db.port = ""
	}

	if err = db.prepare(); err != nil {
		return
	}

	err = db.dump()
	return
}

func (db *PostgreSQL) prepare() (err error) {
	// pg_dump command
	var dumpArgs []string
	if len(db.database) == 0 {
		return fmt.Errorf("PostgreSQL database config is required")
	}
	if len(db.host) > 0 {
		dumpArgs = append(dumpArgs, "--host="+db.host)
	}
	if len(db.port) > 0 {
		dumpArgs = append(dumpArgs, "--port="+db.port)
	}
	if len(db.socket) > 0 {
		host := filepath.Dir(db.socket)
		port := strings.TrimPrefix(filepath.Ext(db.socket), ".")
		dumpArgs = append(dumpArgs, "--host="+host, "--port="+port)
	}
	if len(db.username) > 0 {
		dumpArgs = append(dumpArgs, "--username="+db.username)
	}
	if len(db.args) > 0 {
		dumpArgs = append(dumpArgs, db.args)
	}

	db.dumpCommand = "pg_dump " + strings.Join(dumpArgs, " ") + " " + db.database

	return nil
}

func (db *PostgreSQL) dump() error {
	logger := logger.Tag("PostgreSQL")

	dumpFilePath := path.Join(db.dumpPath, db.database+".sql")
	logger.Info("-> Dumping PostgreSQL...")
	if len(db.password) > 0 {
		os.Setenv("PGPASSWORD", db.password)
	}
	_, err := helper.Exec(db.dumpCommand, "-f", dumpFilePath)
	if err != nil {
		return err
	}
	logger.Info("dump path:", dumpFilePath)
	return nil
}
