package database

import (
	"fmt"
	"path"
	"sort"
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
// separate_files: true (with all_databases, dump each database into its own file)
// exclude_databases: (with separate_files, database names to skip)
// username: root
// password:
// args:
type MySQL struct {
	Base
	host             string
	port             string
	socket           string
	database         string
	username         string
	password         string
	tables           []string
	excludeTables    []string
	args             string
	allDatabases     bool
	separateFiles    bool
	excludeDatabases []string
}

var (
	mysqldumpCli = "mysqldump"
	mysqlCli     = "mysql"

	// Databases that `mysqldump --all-databases` never dumps, they are always
	// skipped when dumping each database into its own file.
	mysqlSystemDatabases = []string{"information_schema", "performance_schema"}
)

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
	db.separateFiles = viper.GetBool("separate_files")
	db.excludeDatabases = viper.GetStringSlice("exclude_databases")

	// tables/exclude_tables options are not compatible with all databases mode
	if db.allDatabases && (len(db.tables) > 0 || len(db.excludeTables) > 0) {
		return fmt.Errorf("tables and exclude_tables options are not supported when using all_databases: true")
	}

	if db.separateFiles && !db.allDatabases {
		return fmt.Errorf("separate_files option is only supported when using all_databases: true")
	}

	// socket
	if len(db.socket) != 0 {
		db.host = ""
		db.port = ""
	}

	return nil
}

// connectionArgs are the options shared by every command we run against the server
func (db *MySQL) connectionArgs() []string {
	args := []string{}
	if len(db.host) > 0 {
		args = append(args, "--host", db.host)
	}
	if len(db.port) > 0 {
		args = append(args, "--port", db.port)
	}
	if len(db.socket) > 0 {
		args = append(args, "--socket", db.socket)
	}
	if len(db.username) > 0 {
		args = append(args, "-u", db.username)
	}
	if len(db.password) > 0 {
		args = append(args, `-p`+db.password)
	}

	return args
}

func (db *MySQL) build() string {
	if db.allDatabases && !db.separateFiles {
		dumpArgs := append(db.connectionArgs(), "--all-databases")
		if len(db.args) > 0 {
			dumpArgs = append(dumpArgs, db.args)
		}
		dumpArgs = append(dumpArgs, "--result-file="+path.Join(db.dumpPath, "all-databases.sql"))

		return mysqldumpCli + " " + strings.Join(dumpArgs, " ")
	}

	return db.buildDatabase(db.database)
}

// buildDatabase builds the dump command of a single database, the dump is
// written to a file named after the database.
func (db *MySQL) buildDatabase(database string) string {
	dumpArgs := db.connectionArgs()

	for _, table := range db.excludeTables {
		dumpArgs = append(dumpArgs, "--ignore-table="+database+"."+table)
	}

	if len(db.args) > 0 {
		dumpArgs = append(dumpArgs, db.args)
	}

	dumpArgs = append(dumpArgs, database)
	dumpArgs = append(dumpArgs, db.tables...)
	dumpArgs = append(dumpArgs, "--result-file="+path.Join(db.dumpPath, database+".sql"))

	return mysqldumpCli + " " + strings.Join(dumpArgs, " ")
}

// listDatabasesArgs builds the `mysql` arguments used to read the database names
func (db *MySQL) listDatabasesArgs() []string {
	return append(db.connectionArgs(), "--skip-column-names", "--batch", "-e", "SHOW DATABASES")
}

// listDatabases returns the databases to dump, system and excluded ones are removed
func (db *MySQL) listDatabases() ([]string, error) {
	output, err := helper.Exec(mysqlCli, db.listDatabasesArgs()...)
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

func (db *MySQL) shouldSkipDatabase(database string) bool {
	for _, name := range mysqlSystemDatabases {
		if strings.EqualFold(database, name) {
			return true
		}
	}

	for _, name := range db.excludeDatabases {
		if strings.EqualFold(database, name) {
			return true
		}
	}

	return false
}

func (db *MySQL) perform() error {
	logger := logger.Tag("MySQL")

	logger.Info("-> Dumping MySQL...")

	if db.allDatabases && db.separateFiles {
		databases, err := db.listDatabases()
		if err != nil {
			return fmt.Errorf("-> Dump error: %s", err)
		}

		if len(databases) == 0 {
			logger.Warn("No databases to dump")
			return nil
		}

		logger.Infof("-> Dumping %d database(s) into separate files: %v", len(databases), databases)
		for _, database := range databases {
			logger.Info("-> Dumping database:", database)
			if _, err := helper.Exec(db.buildDatabase(database)); err != nil {
				return fmt.Errorf("-> Dump error on database %s: %s", database, err)
			}
		}

		logger.Info("dump path:", db.dumpPath)
		return nil
	}

	_, err := helper.Exec(db.build())
	if err != nil {
		return fmt.Errorf("-> Dump error: %s", err)
	}
	logger.Info("dump path:", db.dumpPath)
	return nil
}
