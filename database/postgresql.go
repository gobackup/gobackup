package database

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
)

// PostgreSQL database
//
// ref:
// https://www.postgresql.org/docs/current/app-pgdump.html
//
// # Keys
//
//   - type: postgresql
//   - host: localhost
//   - port: 5432
//   - socket:
//   - database:
//   - username:
//   - password:
//   - tables:
//   - exclude_tables:
//   - args:
type PostgreSQL struct {
	Base
	host          string
	port          string
	socket        string
	database      string
	username      string
	tables        []string
	excludeTables []string
	password      string
	compress      string
	format        string
	args          string
	_dumpFilePath string
}

var (
	PostgreSQLCompressionExt = map[string]string{
		"gzip": "gz",
		"lz4":  "lz4",
		"zstd": "zst",
	}
)

func (db *PostgreSQL) init() (err error) {
	viper := db.viper
	viper.SetDefault("host", "localhost")
	viper.SetDefault("port", 5432)

	db.host = viper.GetString("host")
	db.port = viper.GetString("port")
	db.socket = viper.GetString("socket")
	db.database = viper.GetString("database")
	db.username = viper.GetString("username")
	db.password = viper.GetString("password")
	db.tables = viper.GetStringSlice("tables")
	db.excludeTables = viper.GetStringSlice("exclude_tables")
	db.compress = viper.GetString("compress")
	db.format = ".sql"
	db.args = viper.GetString("args")

	if len(db.database) == 0 {
		return fmt.Errorf("PostgreSQL database config is required")
	}

	if len(db.compress) > 0 {
		compression := strings.Split(db.compress, ":")[0]
		if _, ok := PostgreSQLCompressionExt[compression]; !ok {
			return fmt.Errorf("PostgreSQL compression type is not allowed: %s", compression)
		}
		db.format = fmt.Sprintf("%s.%s", db.format, PostgreSQLCompressionExt[compression])
	}

	db._dumpFilePath = path.Join(db.dumpPath, db.database+db.format)

	// socket
	if len(db.socket) != 0 {
		db.host = ""
		db.port = ""
	}

	return nil
}

func (db *PostgreSQL) build() string {
	// pg_dump command
	var dumpArgs []string

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

	// include / exclude tables
	if len(db.tables) > 0 {
		dumpArgs = append(dumpArgs, "--table="+strings.Join(db.tables, " --table="))
	}

	if len(db.excludeTables) > 0 {
		dumpArgs = append(dumpArgs, "--exclude-table="+strings.Join(db.excludeTables, " --exclude-table="))
	}

	if len(db.compress) > 0 {
		dumpArgs = append(dumpArgs, "--compress="+db.compress, "--format=custom")
	}

	if len(db.args) > 0 {
		dumpArgs = append(dumpArgs, db.args)
	}

	dumpArgs = append(dumpArgs, db.database)
	dumpArgs = append(dumpArgs, "-f", db._dumpFilePath)

	return "pg_dump " + strings.Join(dumpArgs, " ")
}

func (db *PostgreSQL) perform() error {
	logger := logger.Tag("PostgreSQL")

	logger.Info("-> Dumping PostgreSQL...")
	if len(db.password) > 0 {
		os.Setenv("PGPASSWORD", db.password)
	}

	_, err := helper.Exec(db.build())
	if err != nil {
		return err
	}
	logger.Info("dump path:", db._dumpFilePath)
	return nil
}
