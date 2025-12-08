package database

import (
	"fmt"
	"strings"

	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
)

// MSSQL database
//
// type: mssql
// host: 127.0.0.1
// port: 1433
// database: [string]
// username: [string]
// password: [string]
// trustServerCertificate: [true, false]
// backupAllDatases: [true, false] # if true, backup all databases excluding system databases and ignores database parameter
// skipDatabases: [array] # list of database names to skip when backupAllDatases is true
// args:
type MSSQL struct {
	Base
	host                   string
	port                   string
	database               string
	username               string
	password               string
	trustServerCertificate bool
	includeAllDatabases    bool
	skipDatabases          []string
	args                   string
}

var (
	sqlpackageCli = "sqlpackage"
)

func (db *MSSQL) init() (err error) {
	viper := db.viper
	viper.SetDefault("trustServerCertificate", false)
	viper.SetDefault("host", "127.0.0.1")
	viper.SetDefault("port", 1433)
	viper.SetDefault("username", "sa")
	viper.SetDefault("backupAllDatases", false)

	db.host = viper.GetString("host")
	db.port = viper.GetString("port")
	db.database = viper.GetString("database")
	db.username = viper.GetString("username")
	db.password = viper.GetString("password")
	db.trustServerCertificate = viper.GetBool("trustServerCertificate")
	db.includeAllDatabases = viper.GetBool("backupAllDatases")
	db.skipDatabases = viper.GetStringSlice("skipDatabases")
	db.args = viper.GetString("args")

	return nil
}

func (db *MSSQL) build() string {
	return sqlpackageCli + " " +
		"/Action:Export " +
		db.nameOption() + " " +
		db.credentialOptions() + " " +
		db.connectivityOptions() + " " +
		db.additionOption() + " " +
		"/TargetFile:" + db.dumpPath + "/" + db.database + ".bacpac"
}

func (db *MSSQL) nameOption() string {
	return "/SourceDatabaseName:" + db.database
}

func (db *MSSQL) credentialOptions() string {
	opts := []string{}
	if len(db.username) > 0 {
		opts = append(opts, "/SourceUser:"+db.username)
	}
	if len(db.password) > 0 {
		opts = append(opts, "/SourcePassword:"+db.password)
	}
	return strings.Join(opts, " ")
}

func (db *MSSQL) connectivityOptions() string {
	var host = db.host
	var port = db.port

	if len(host) == 0 {
		host = "127.0.0.1"
	}
	if len(port) == 0 {
		port = "1433"
	}

	return "/SourceServerName:" + host + "," + port
}

func (db *MSSQL) additionOption() string {
	opts := []string{}
	if db.trustServerCertificate {
		opts = append(opts, "/SourceTrustServerCertificate:True")
	}

	if len(db.args) > 0 {
		opts = append(opts, db.args)
	}

	return strings.Join(opts, " ")
}

func (db *MSSQL) getAllDatabases() ([]string, error) {
	// Exclude system databases
	query := "SET NOCOUNT ON; SELECT name FROM sys.databases WHERE name NOT IN ('master', 'tempdb', 'model', 'msdb')"
	args := []string{
		"-S", db.host + "," + db.port,
		"-Q", query,
		"-h-1",
		"-W",
	}

	if len(db.username) > 0 {
		args = append(args, "-U", db.username)
	}
	if len(db.password) > 0 {
		args = append(args, "-P", db.password)
	}
	if db.trustServerCertificate {
		args = append(args, "-C")
	}

	output, err := helper.Exec("sqlcmd", args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get databases: %s", err)
	}

	lines := strings.Split(strings.TrimSpace(output), "\n")
	databases := []string{}
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if len(line) > 0 {
			databases = append(databases, line)
		}
	}

	return databases, nil
}

func (db *MSSQL) shouldSkipDatabase(databaseName string) bool {
	if !db.includeAllDatabases {
		return false
	}
	for _, skipDB := range db.skipDatabases {
		if strings.EqualFold(databaseName, skipDB) {
			return true
		}
	}
	return false
}

func (db *MSSQL) perform() error {
	logger := logger.Tag("MSSQL")

	if db.includeAllDatabases {
		databases, err := db.getAllDatabases()
		if err != nil {
			return fmt.Errorf("-> Failed to get databases: %s", err)
		}

		if len(databases) == 0 {
			logger.Warn("No non-system databases found")
			return nil
		}

		filteredDatabases := []string{}
		for _, databaseName := range databases {
			if !db.shouldSkipDatabase(databaseName) {
				filteredDatabases = append(filteredDatabases, databaseName)
			} else {
				logger.Infof("Skipping database: %s", databaseName)
			}
		}

		if len(filteredDatabases) == 0 {
			logger.Warn("No databases to backup after filtering")
			return nil
		}

		logger.Infof("Found %d database(s) to backup", len(filteredDatabases))
		logger.Infof("Databases: %v", filteredDatabases)
		for _, databaseName := range filteredDatabases {
			// Update the database field directly so build() uses the correct database name
			db.database = databaseName
			logger.Infof("Backing up database: %s", databaseName)
			out, err := helper.Exec(db.build())
			if err != nil {
				return fmt.Errorf("-> Dump error for database %s: %s", databaseName, err)
			}
			logger.Info(out)
			logger.Info("dump path:", db.dumpPath)
		}
		return nil
	}

	if len(db.database) == 0 {
		return fmt.Errorf("database config is required when backupAllDatases is false")
	}

	out, err := helper.Exec(db.build())
	if err != nil {
		return fmt.Errorf("-> Dump error: %s", err)
	}
	logger.Info(out)
	logger.Info("dump path:", db.dumpPath)
	return nil
}
