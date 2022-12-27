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
// host: 127.0.0.1
// port: 27017
// database:
// username:
// password:
// authdb:
// collection:
// gzip:
// oplog: false
type MongoDB struct {
	Base
	host     string
	port     string
	database string
	username string
	password string
	authdb   string
	oplog    bool
	args     string
}

var (
	mongodumpCli = "mongodump"
)

func (db *MongoDB) perform() (err error) {
	viper := db.viper
	viper.SetDefault("oplog", false)
	viper.SetDefault("host", "127.0.0.1")
	viper.SetDefault("port", 27017)

	db.host = viper.GetString("host")
	db.port = viper.GetString("port")
	db.database = viper.GetString("database")
	db.username = viper.GetString("username")
	db.password = viper.GetString("password")
	db.oplog = viper.GetBool("oplog")
	db.authdb = viper.GetString("authdb")
	db.args = viper.GetString("args")

	err = db.dump()
	if err != nil {
		return err
	}
	return nil
}

func (db *MongoDB) mongodump() string {
	return mongodumpCli + " " +
		db.nameOption() + " " +
		db.credentialOptions() + " " +
		db.connectivityOptions() + " " +
		db.oplogOption() + " " +
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
	if len(db.args) > 0 {
		dumpArgs = append(dumpArgs, db.args)
	}

	return strings.Join(opts, " ")
}

func (db *MongoDB) oplogOption() string {
	if db.oplog {
		return "--oplog"
	}

	return ""
}

func (db *MongoDB) dump() error {
	logger := logger.Tag("MongoDB")

	out, err := helper.Exec(db.mongodump())
	if err != nil {
		return fmt.Errorf("-> Dump error: %s", err)
	}
	logger.Info(out)
	logger.Info("dump path:", db.dumpPath)
	return nil
}
