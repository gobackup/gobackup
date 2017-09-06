package database

import (
	"github.com/huacnlee/gobackup/logger"
)

// MySQL database
type MySQL struct {
}

func (ctx MySQL) Perform() (err error) {
	logger.Println("Perform MySQL...")
	return
}
