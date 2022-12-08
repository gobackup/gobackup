package database

import (
	"fmt"
	"os"
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
	logger   *logger.Logger
}

func (db *SQLite) perform() error {
	viper := db.viper

	db.path = helper.ExplandHome(viper.GetString("path"))
	db.database = strings.TrimSuffix(filepath.Base(db.path), filepath.Ext(db.path))

	return db.dump()
}

func (db *SQLite) dump() error {
	logger := db.logger

	dumpFilePath := filepath.Join(db.dumpPath, db.database+".sql")
	logger.Info("-> Dumping SQLite...")
	if out, err := helper.Exec(fmt.Sprintf("sqlite3 %s .dump", db.path)); err != nil {
		return err
	} else if err := os.WriteFile(dumpFilePath, []byte(out), 0644); err != nil {
		return err
	}
	logger.Info("dump path:", dumpFilePath)
	return nil
}
