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
// database:
// username:
// password:
// trustServerCertificate:
// args:
type MSSQL struct {
	Base
	host                   string
	port                   string
	database               string
	username               string
	password               string
	trustServerCertificate bool
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

	db.host = viper.GetString("host")
	db.port = viper.GetString("port")
	db.database = viper.GetString("database")
	db.username = viper.GetString("username")
	db.password = viper.GetString("password")
	db.trustServerCertificate = viper.GetBool("trustServerCertificate")
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

func (db *MSSQL) perform() error {
	logger := logger.Tag("MSSQL")

	out, err := helper.Exec(db.build())
	if err != nil {
		return fmt.Errorf("-> Dump error: %s", err)
	}
	logger.Info(out)
	logger.Info("dump path:", db.dumpPath)
	return nil
}
