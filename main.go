package main

import (
	"github.com/huacnlee/gobackup/database"
	"github.com/huacnlee/gobackup/logger"
)

func main() {
	db, err := database.New("redis")
	if err != nil {
		logger.Error(err)
	}
	err = db.Perform()
	if err != nil {
		logger.Error(err)
	}
}
