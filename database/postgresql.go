package database

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"sort"
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
//   - all_databases: false
//   - separate_files: false
//   - exclude_databases:
type PostgreSQL struct {
	Base
	host             string
	port             string
	socket           string
	database         string
	username         string
	tables           []string
	excludeTables    []string
	password         string
	compress         string
	format           string
	args             string
	allDatabases     bool
	separateFiles    bool
	excludeDatabases []string
	_dumpFilePath    string
}

var (
	PostgreSQLCompressionExt = map[string]string{
		"gzip": "gz",
		"lz4":  "lz4",
		"zstd": "zst",
	}

	// Database used to connect to the server when reading the database names
	postgreSQLDefaultDatabase = "postgres"

	// File holding the cluster wide objects (roles, tablespaces) when each
	// database is dumped into its own file
	postgreSQLGlobalsFileName = "globals.sql"
)

func (db *PostgreSQL) init() (err error) {
	viper := db.viper
	viper.SetDefault("host", "localhost")
	viper.SetDefault("port", 5432)
	viper.SetDefault("all_databases", false)
	viper.SetDefault("separate_files", false)

	db.host = viper.GetString("host")
	db.port = viper.GetString("port")
	db.socket = viper.GetString("socket")
	db.database = viper.GetString("database")
	db.username = viper.GetString("username")
	db.password = viper.GetString("password")
	db.allDatabases = viper.GetBool("all_databases")
	db.separateFiles = viper.GetBool("separate_files")
	db.excludeDatabases = viper.GetStringSlice("exclude_databases")
	db.tables = viper.GetStringSlice("tables")
	db.excludeTables = viper.GetStringSlice("exclude_tables")
	db.compress = viper.GetString("compress")
	db.format = ".sql"
	db.args = viper.GetString("args")

	if !db.allDatabases && len(db.database) == 0 {
		return fmt.Errorf("PostgreSQL database config is required")
	}

	if db.separateFiles && !db.allDatabases {
		return fmt.Errorf("separate_files option is only supported when using all_databases: true")
	}

	if len(db.compress) > 0 {
		compression := strings.Split(db.compress, ":")[0]
		if _, ok := PostgreSQLCompressionExt[compression]; !ok {
			return fmt.Errorf("PostgreSQL compression type is not allowed: %s", compression)
		}
		db.format = fmt.Sprintf("%s.%s", db.format, PostgreSQLCompressionExt[compression])
	}

	// In separate files mode the dump path is resolved per database
	if !db.separateFiles {
		if db.allDatabases {
			db._dumpFilePath = db.dumpFilePath("all_databases")
		} else {
			db._dumpFilePath = db.dumpFilePath(db.database)
		}
	}

	// socket
	if len(db.socket) != 0 {
		db.host = ""
		db.port = ""
	}

	return nil
}

func (db *PostgreSQL) dumpFilePath(database string) string {
	return path.Join(db.dumpPath, database+db.format)
}

// connectionArgs are the flags shared by pg_dump, pg_dumpall and psql
func (db *PostgreSQL) connectionArgs() []string {
	args := []string{}
	if len(db.host) > 0 {
		args = append(args, "--host="+db.host)
	}
	if len(db.port) > 0 {
		args = append(args, "--port="+db.port)
	}
	if len(db.socket) > 0 {
		host := filepath.Dir(db.socket)
		port := strings.TrimPrefix(filepath.Ext(db.socket), ".")
		args = append(args, "--host="+host, "--port="+port)
	}
	if len(db.username) > 0 {
		args = append(args, "--username="+db.username)
	}

	return args
}

func (db *PostgreSQL) build() string {
	if db.allDatabases && !db.separateFiles {
		// pg_dumpall command for all databases
		pgDumpallArgs := db.connectionArgs()

		if len(db.args) > 0 {
			pgDumpallArgs = append(pgDumpallArgs, db.args)
		}

		// Build the complete pg_dumpall command with output redirection
		return "pg_dumpall " + strings.Join(pgDumpallArgs, " ") + " > " + db._dumpFilePath
	}

	return db.buildDatabase(db.database, db._dumpFilePath)
}

// buildDatabase builds the pg_dump command of a single database
func (db *PostgreSQL) buildDatabase(database, dumpFilePath string) string {
	dumpArgs := db.connectionArgs()

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

	dumpArgs = append(dumpArgs, database)
	dumpArgs = append(dumpArgs, "-f", dumpFilePath)

	return "pg_dump " + strings.Join(dumpArgs, " ")
}

// buildGlobals builds the pg_dumpall command that dumps the cluster wide
// objects, they are not part of a single database dump but are required to
// restore it.
func (db *PostgreSQL) buildGlobals() string {
	dumpArgs := append(db.connectionArgs(), "--globals-only")
	dumpArgs = append(dumpArgs, "-f", path.Join(db.dumpPath, postgreSQLGlobalsFileName))

	return "pg_dumpall " + strings.Join(dumpArgs, " ")
}

// listDatabasesArgs builds the `psql` arguments used to read the database names
func (db *PostgreSQL) listDatabasesArgs() []string {
	database := db.database
	if len(database) == 0 {
		database = postgreSQLDefaultDatabase
	}

	return append(db.connectionArgs(),
		"--no-align",
		"--tuples-only",
		"--dbname="+database,
		"--command=SELECT datname FROM pg_database WHERE datallowconn AND NOT datistemplate",
	)
}

// listDatabases returns the databases to dump, excluded ones are removed
func (db *PostgreSQL) listDatabases() ([]string, error) {
	output, err := helper.Exec("psql", db.listDatabasesArgs()...)
	if err != nil {
		return nil, fmt.Errorf("failed to list databases: %s", err)
	}

	databases := []string{}
	for _, line := range strings.Split(output, "\n") {
		database := strings.TrimSpace(line)
		if len(database) == 0 || db.shouldSkipDatabase(database) {
			continue
		}

		databases = append(databases, database)
	}
	sort.Strings(databases)

	return databases, nil
}

func (db *PostgreSQL) shouldSkipDatabase(database string) bool {
	for _, name := range db.excludeDatabases {
		if strings.EqualFold(database, name) {
			return true
		}
	}

	return false
}

func (db *PostgreSQL) perform() error {
	logger := logger.Tag("PostgreSQL")

	logger.Info("-> Dumping PostgreSQL...")
	if len(db.password) > 0 {
		os.Setenv("PGPASSWORD", db.password)
	}

	if db.allDatabases && db.separateFiles {
		return db.performSeparateFiles()
	}

	var err error
	if db.allDatabases {
		// Use ExecScript for pg_dumpall to properly handle shell redirection
		_, err = helper.ExecScript(db.build())
	} else {
		_, err = helper.Exec(db.build())
	}
	if err != nil {
		return err
	}
	logger.Info("dump path:", db._dumpFilePath)
	return nil
}

// performSeparateFiles dumps the cluster wide objects and then each database
// into its own file.
func (db *PostgreSQL) performSeparateFiles() error {
	logger := logger.Tag("PostgreSQL")

	databases, err := db.listDatabases()
	if err != nil {
		return err
	}

	if len(databases) == 0 {
		logger.Warn("No databases to dump")
		return nil
	}

	logger.Info("-> Dumping globals...")
	if _, err := helper.Exec(db.buildGlobals()); err != nil {
		return fmt.Errorf("dump globals error: %s", err)
	}

	logger.Infof("-> Dumping %d database(s) into separate files: %v", len(databases), databases)
	for _, database := range databases {
		logger.Info("-> Dumping database:", database)
		if _, err := helper.Exec(db.buildDatabase(database, db.dumpFilePath(database))); err != nil {
			return fmt.Errorf("dump error on database %s: %s", database, err)
		}
	}

	logger.Info("dump path:", db.dumpPath)
	return nil
}
