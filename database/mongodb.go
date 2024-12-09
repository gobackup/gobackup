package database

import (
	"fmt"
	"strings"

	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
)

// MongoDB database
//
// type: mongodb
// uri: mongodb://username:password@host:port/database?authSource=database
// host: 127.0.0.1
// port: 27017
// database:
// username:
// password:
// authdb:
// exclude_tables:
// exclude_tables_prefix:
// oplog: false
// args:
type MongoDB struct {
	Base
	uri                 string
	host                string
	port                string
	database            string
	username            string
	password            string
	authdb              string
	excludeTables       []string
	excludeTablesPrefix []string
	oplog               bool
	args                string
}

var (
	mongodumpCli = "mongodump"
)

func (db *MongoDB) init() (err error) {
	viper := db.viper
	viper.SetDefault("oplog", false)
	viper.SetDefault("host", "127.0.0.1")
	viper.SetDefault("port", 27017)

	db.uri = viper.GetString("uri")
	db.host = viper.GetString("host")
	db.port = viper.GetString("port")
	db.database = viper.GetString("database")
	db.username = viper.GetString("username")
	db.password = viper.GetString("password")
	db.oplog = viper.GetBool("oplog")
	db.authdb = viper.GetString("authdb")
	db.excludeTables = viper.GetStringSlice("exclude_tables")
	db.excludeTablesPrefix = viper.GetStringSlice("exclude_tables_prefix")
	db.args = viper.GetString("args")

	return nil
}

func (db *MongoDB) build() string {
	if len(db.uri) > 0 {
		return mongodumpCli + " " +
			"--uri=" + db.uri + " " +
			db.additionOption() + " " +
			"--out=" + db.dumpPath
	}
	return mongodumpCli + " " +
		db.nameOption() + " " +
		db.credentialOptions() + " " +
		db.connectivityOptions() + " " +
		db.additionOption() + " " +
		"--out=" + db.dumpPath
}

func (db *MongoDB) nameOption() string {
	return "--db=" + db.database
}

func (db *MongoDB) credentialOptions() string {
	opts := []string{}
	if len(db.username) > 0 {
		opts = append(opts, "--username="+db.username)
	}
	if len(db.password) > 0 {
		opts = append(opts, `--password=`+db.password)
	}
	if len(db.authdb) > 0 {
		opts = append(opts, "--authenticationDatabase="+db.authdb)
	}
	return strings.Join(opts, " ")
}

func (db *MongoDB) connectivityOptions() string {
	opts := []string{}
	if len(db.host) > 0 {
		opts = append(opts, "--host="+db.host+"")
	}
	if len(db.port) > 0 {
		opts = append(opts, "--port="+db.port+"")
	}

	return strings.Join(opts, " ")
}

func (db *MongoDB) additionOption() string {
	opts := []string{}
	if db.oplog {
		opts = append(opts, "--oplog")
	}

	for _, table := range db.excludeTables {
		opts = append(opts, "--excludeCollection="+table)
	}

	for _, table := range db.excludeTablesPrefix {
		opts = append(opts, "--excludeCollectionsWithPrefix="+table)
	}

	if len(db.args) > 0 {
		opts = append(opts, db.args)
	}

	return strings.Join(opts, " ")
}

func (db *MongoDB) perform() error {
	logger := logger.Tag("MongoDB")

	out, err := helper.Exec(db.build())
	if err != nil {
		return fmt.Errorf("-> Dump error: %s", err)
	}
	logger.Info(out)
	logger.Info("dump path:", db.dumpPath)
	return nil
}
