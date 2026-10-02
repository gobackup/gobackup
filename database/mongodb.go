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
// all_databases: false
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
	allDatabases        bool
}

var (
	mongodumpCli = "mongodump"
)

func (db *MongoDB) init() (err error) {
	viper := db.viper
	viper.SetDefault("oplog", false)
	viper.SetDefault("host", "127.0.0.1")
	viper.SetDefault("port", 27017)
	viper.SetDefault("all_databases", false)

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
	db.allDatabases = viper.GetBool("all_databases")

	if !db.allDatabases && len(db.uri) == 0 && len(db.database) == 0 {
		return fmt.Errorf("MongoDB database config is required")
	}

	return nil
}

// dumpCommand builds the mongodump command without any output target.
func (db *MongoDB) dumpCommand() string {
	if len(db.uri) > 0 {
		cmd := mongodumpCli + " " + "--uri=" + db.uri
		if opts := db.additionOption(); len(opts) > 0 {
			cmd += " " + opts
		}
		return cmd
	}

	cmd := mongodumpCli + " "
	if !db.allDatabases {
		cmd += db.nameOption() + " "
	}
	cmd += db.credentialOptions() + " " + db.connectivityOptions()
	if opts := db.additionOption(); len(opts) > 0 {
		cmd += " " + opts
	}

	return cmd
}

func (db *MongoDB) build() string {
	return db.dumpCommand() + " " + "--out=" + db.dumpPath
}

// stream starts mongodump writing an archive to stdout.
// Restore with: mongorestore --archive < file
func (db *MongoDB) stream() (*Stream, error) {
	name := db.database
	if db.allDatabases || len(name) == 0 {
		name = "all-databases"
	}
	return startStream(db.dumpCommand()+" --archive", name+".archive", ".archive")
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
