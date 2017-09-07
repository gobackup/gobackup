package database

import (
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/logger"
)

// MySQL database
type MySQL struct {
	Name     string
	Host     string
	Port     int
	Database string
	Username string
	Password string
}

// NewMySQL instrance
func newMySQL(dbCfg config.SubConfig) (ctx MySQL) {
	viper := dbCfg.Viper
	viper.SetDefault("host", "localhost")
	viper.SetDefault("username", "root")
	viper.SetDefault("port", 3306)

	ctx = MySQL{
		Name:     dbCfg.Name,
		Host:     viper.GetString("host"),
		Port:     viper.GetInt("port"),
		Database: viper.GetString("database"),
		Username: viper.GetString("username"),
		Password: viper.GetString("password"),
	}

	return ctx
}

func (ctx MySQL) perform() (err error) {
	logger.Warn("Not implement")
	return
}

func (ctx MySQL) dump() (err error) {
	return
}
