package database

import (
	"fmt"
	"strings"

	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
)

// Atlas database
//
// type: atlas
// uri:
// exclude_tables:
// args:

type Atlas struct {
	Base
	uri           string
	excludeTables []string
	args          string
}

func (db *Atlas) init() (err error) {
	viper := db.viper

	db.uri = viper.GetString("uri")
	db.excludeTables = viper.GetStringSlice("exclude_tables")
	db.args = viper.GetString("args")
	return nil
}

func (db *Atlas) build() string {
	return mongodumpCli + " " +
		"--uri=" + db.uri + " " +
		db.additionOption() + " " +
		"--out=" + db.dumpPath
}

func (db *Atlas) additionOption() string {
	opts := []string{}
	for _, table := range db.excludeTables {
		opts = append(opts, "--excludeCollection="+table)
	}

	if len(db.args) > 0 {
		opts = append(opts, db.args)
	}

	return strings.Join(opts, " ")
}

func (db *Atlas) perform() error {
	logger := logger.Tag("MongoDB")

	out, err := helper.Exec(db.build())
	if err != nil {
		return fmt.Errorf("-> Dump error: %s", err)
	}
	logger.Info(out)
	logger.Info("dump path:", db.dumpPath)
	return nil
}
