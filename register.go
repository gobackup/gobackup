package main

import (
	"github.com/gobackup/gobackup/database"
	"github.com/gobackup/gobackup/notifier"
	"github.com/gobackup/gobackup/storage"
)

func init() {
	// Register all built-in database types
	database.RegisterAll()

	// Register all built-in storage types
	storage.RegisterAll()

	// Register all built-in notifier types
	notifier.RegisterAll()
}
