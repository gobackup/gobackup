package database

import (
	"fmt"
	"os"
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
// database: <database_name>
// all_databases: true (backup all databases)
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
	allDatabases  bool
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

	db.allDatabases = viper.GetBool("all_databases")

	// tables/exclude_tables options are not compatible with all databases mode
	if db.allDatabases && (len(db.tables) > 0 || len(db.excludeTables) > 0) {
		return fmt.Errorf("tables and exclude_tables options are not supported when using all_databases: true")
	}

	// socket
	if len(db.socket) != 0 {
		db.host = ""
		db.port = ""
	}

	return nil
}

func (db *MySQL) build() string {
	var dumpArgs []string
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
	// The password is passed via a temporary --defaults-extra-file in perform(),
	// never on the command line, so special characters in it need no escaping.

	// Handle all databases mode
	if db.allDatabases {
		dumpArgs = append(dumpArgs, "--all-databases")
	} else {
		// Single database mode with optional table filtering
		for _, table := range db.excludeTables {
			dumpArgs = append(dumpArgs, "--ignore-table="+db.database+"."+table)
		}
	}

	if len(db.args) > 0 {
		dumpArgs = append(dumpArgs, db.args)
	}

	// Add database name and tables for single database mode
	if !db.allDatabases {
		dumpArgs = append(dumpArgs, db.database)
		if len(db.tables) > 0 {
			dumpArgs = append(dumpArgs, db.tables...)
		}
	}

	// Determine dump file name
	dumpFileName := db.database + ".sql"
	if db.allDatabases {
		dumpFileName = "all-databases.sql"
	}
	dumpFilePath := path.Join(db.dumpPath, dumpFileName)
	dumpArgs = append(dumpArgs, "--result-file="+dumpFilePath)

	return "mysqldump" + " " + strings.Join(dumpArgs, " ")
}

func (db *MySQL) perform() error {
	logger := logger.Tag("MySQL")

	logger.Info("-> Dumping MySQL...")

	command := db.build()
	if len(db.password) > 0 {
		// Pass the password through a temporary defaults-extra-file rather than the
		// command line or the deprecated MYSQL_PWD env var. --defaults-extra-file must
		// be the first argument, and the file is removed as soon as the dump finishes.
		confPath, err := db.writePasswordConfig()
		if err != nil {
			return fmt.Errorf("-> Dump error: %s", err)
		}
		defer os.Remove(confPath)
		command = strings.Replace(command, "mysqldump", "mysqldump --defaults-extra-file="+confPath, 1)
	}

	_, err := helper.Exec(command)
	if err != nil {
		return fmt.Errorf("-> Dump error: %s", err)
	}

	logger.Info("dump path:", db.dumpPath)
	return nil
}

// writePasswordConfig writes the MySQL password into a temporary option file
// (created with 0600 permissions) and returns its path, for use with
// --defaults-extra-file. This keeps the password off the command line (invisible in
// `ps`) and out of the process environment, without relying on the deprecated
// MYSQL_PWD variable. The value is double-quoted with backslashes and double quotes
// escaped, so any special character in the password is handled.
func (db *MySQL) writePasswordConfig() (string, error) {
	f, err := os.CreateTemp("", "gobackup-mysql-*.cnf")
	if err != nil {
		return "", err
	}
	defer f.Close()

	escaped := strings.NewReplacer(`\`, `\\`, `"`, `\"`).Replace(db.password)
	content := "[client]\npassword=\"" + escaped + "\"\n"
	if _, err := f.WriteString(content); err != nil {
		os.Remove(f.Name())
		return "", err
	}

	return f.Name(), nil
}
