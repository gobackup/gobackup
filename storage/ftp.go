package storage

import (
	"os"
	"path"
	"path/filepath"
	// "crypto/tls"
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/logger"
	"github.com/secsy/goftp"
)

// FTP storage
type FTP struct {
	path     string
	host     string
	port     string
	username string
	password string
}

func (ctx *FTP) perform(model config.ModelConfig, archivePath string) error {
	logger.Info("=> storage | FTP")

	model.StoreWith.Viper.SetDefault("port", "21")

	ctx.host = model.StoreWith.Viper.GetString("host")
	ctx.port = model.StoreWith.Viper.GetString("port")
	ctx.path = model.StoreWith.Viper.GetString("path")
	ctx.username = model.StoreWith.Viper.GetString("username")
	ctx.password = model.StoreWith.Viper.GetString("password")

	ftpConfig := goftp.Config{
		User:     model.StoreWith.Viper.GetString("username"),
		Password: model.StoreWith.Viper.GetString("password"),
	}

	ftp, err := goftp.DialConfig(ftpConfig, model.StoreWith.Viper.GetString("host")+":"+model.StoreWith.Viper.GetString("port"))
	if err != nil {
		return err
	}
	defer ftp.Close()

	logger.Info("-> Uploading...")
	_, err = ftp.Stat(ctx.path)
	if os.IsNotExist(err) {
		if _, err := ftp.Mkdir(ctx.path); err != nil {
			return err
		}
	}

	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}

	fileName := filepath.Base(archivePath)
	remotePath := path.Join(ctx.path, fileName)
	logger.Info("-> upload", remotePath)
	err = ftp.Store(remotePath, file)
	if err != nil {
		return err
	}

	logger.Info("Store successed")
	return nil
}
