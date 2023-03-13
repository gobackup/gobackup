package database

import (
	"path/filepath"
	"strings"

	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
)

// SQLite database
//
// type: sqlite
// path:
type SQLite struct {
	Base
	path     string
	database string

	_dumpFilePath string
}

func (db *SQLite) init() error {
	viper := db.viper

	db.path = helper.ExplandHome(viper.GetString("path"))
	db.database = strings.TrimSuffix(filepath.Base(db.path), filepath.Ext(db.path))

	db._dumpFilePath = filepath.Join(db.dumpPath, db.database+".sql")

	return nil
}

func (db *SQLite) build() string {
	args := []string{
		"sqlite3",
		db.path,
		".dump",
		">",
		db._dumpFilePath,
	}
	return strings.Join(args, " ")
}

func (db *SQLite) perform() error {
	logger := logger.Tag("SQLite")

	logger.Info("-> Dumping SQLite...")
	if _, err := helper.Exec(db.build()); err != nil {
		return err
	}

	logger.Info("dump path:", db._dumpFilePath)
	return nil
}
