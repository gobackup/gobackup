package database

import (
	"fmt"
	"path"
	"strings"

	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
)

// Firebird database
//
// With gbak utility : https://www.firebirdsql.org/file/documentation/html/en/firebirddocs/gbak/firebird-gbak.html
//
// type: firebird
// host: 127.0.0.1
// port: 3050
// database:
// username:
// password:
// role:
type Firebird struct {
	Base
	host                   string
	port                   string
	database               string
	username               string
	password               string
	role 				   string
	args                   string
	_dumpFilePath	 	   string
}

func (db *Firebird) init() (err error) {
	viper := db.viper
	viper.SetDefault("host", "127.0.0.1")
	viper.SetDefault("port", 3050)
	viper.SetDefault("username", "sysdba")
	viper.SetDefault("password", "masterkey")

	db.host = viper.GetString("host")
	db.port = viper.GetString("port")
	db.database = viper.GetString("database")
	db.username = viper.GetString("username")
	db.role = viper.GetString("role")
	db.password = viper.GetString("password")
	db.args = viper.GetString("args")

	if len(db.database) == 0 {
		return fmt.Errorf("Firebird database config is required")
	}

	return nil
}

func (db *Firebird) build() string {
	// gbak command
	var gbakArgs []string
	
	gbakArgs = append(gbakArgs, "-b")

	if len(db.username) > 0 {
		gbakArgs = append(gbakArgs, "-user "+db.username)
	}
	if len(db.password) > 0 {
		gbakArgs = append(gbakArgs, "-pass "+db.password)
	}
	if len(db.role) > 0 {
		gbakArgs = append(gbakArgs, "-role "+db.role)
	}

	var dbString string
	if (len(db.host) > 0 && len(db.port) > 0) {
		dbString = db.host+"/"+db.port+":"
	}
	
	dbString = dbString + db.database
	db._dumpFilePath = path.Join(
		db.dumpPath, 
		strings.TrimSuffix(path.Base(db.database), path.Ext(db.database)) + ".fbk",
	)

	gbakArgs = append(gbakArgs, dbString)
	gbakArgs = append(gbakArgs, db._dumpFilePath)

	return "gbak " + strings.Join(gbakArgs, " ")
}

func (db *Firebird) perform() error {
	logger := logger.Tag("Firebird")

	logger.Info("-> Dumping Firebird...")

	_, err := helper.Exec(db.build())
	if err != nil {
		return err
	}
	logger.Info("dump path:", db.dumpPath)
	return nil
}
