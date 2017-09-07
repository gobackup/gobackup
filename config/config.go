package config

import (
	"github.com/huacnlee/gobackup/logger"
	"github.com/spf13/viper"
	"os"
	"path"
)

type Config struct {
	DumpPath  string
	Databases []SubConfig
	Storages  []SubConfig
}

type SubConfig struct {
	Name  string
	Type  string
	Viper *viper.Viper
}

func LoadConfig() (cfg Config) {
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

	cfg = Config{}
	cfg.DumpPath = path.Join(os.TempDir(), "gobackup")
	loadDatabasesConfig(&cfg)
	loadStoragesConfig(&cfg)

	logger.Info(cfg.Databases[0].Type)

	return
}

func loadDatabasesConfig(cfg *Config) {
	subViper := viper.Sub("databases")
	for key := range viper.GetStringMap("databases") {
		dbViper := subViper.Sub(key)
		cfg.Databases = append(cfg.Databases, SubConfig{
			Name:  key,
			Type:  dbViper.GetString("type"),
			Viper: dbViper,
		})
	}
}

func loadStoragesConfig(cfg *Config) {
	subViper := viper.Sub("storages")
	for key := range viper.GetStringMap("storages") {
		dbViper := subViper.Sub(key)
		cfg.Storages = append(cfg.Storages, SubConfig{
			Name:  key,
			Type:  dbViper.GetString("type"),
			Viper: dbViper,
		})
	}
}
