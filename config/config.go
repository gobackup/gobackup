package config

import (
	"github.com/huacnlee/gobackup/logger"
	"github.com/spf13/viper"
	"os"
	"path"
)

var (
	DumpPath  string
	Databases []SubConfig
	Storages  []SubConfig
)

type SubConfig struct {
	Name  string
	Type  string
	Viper *viper.Viper
}

func init() {
	viper.SetConfigType("yaml")
	viper.SetConfigName("gobackup")
	// /etc/gobackup/gobackup.yml
	viper.AddConfigPath("/etc/gobackup/") // path to look for the config file in
	// ~/.gobackup/gobackup.yml
	viper.AddConfigPath("$HOME/.gobackup") // call multiple times to add many search paths
	// ./gobackup.yml
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		logger.Error("Load gobackup config faild", err)
		return
	}

	DumpPath = path.Join(os.TempDir(), "gobackup")
	loadDatabasesConfig()
	loadStoragesConfig()

	return
}

func loadDatabasesConfig() {
	subViper := viper.Sub("databases")
	for key := range viper.GetStringMap("databases") {
		dbViper := subViper.Sub(key)
		Databases = append(Databases, SubConfig{
			Name:  key,
			Type:  dbViper.GetString("type"),
			Viper: dbViper,
		})
	}
}

func loadStoragesConfig() {
	subViper := viper.Sub("storages")
	for key := range viper.GetStringMap("storages") {
		dbViper := subViper.Sub(key)
		Storages = append(Storages, SubConfig{
			Name:  key,
			Type:  dbViper.GetString("type"),
			Viper: dbViper,
		})
	}
}
