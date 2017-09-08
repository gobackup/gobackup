package storage

import (
	// "crypto/tls"
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/logger"
	"gopkg.in/dutchcoders/goftp.v1"
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

	ftp, err := goftp.Connect(ctx.host + ":" + ctx.port)
	if err != nil {
		return err
	}
	defer ftp.Close()

	logger.Info("-> Authorizing FTP...")
	if err := ftp.Login(ctx.username, ctx.password); err != nil {
		return err
	}

	logger.Info("-> Uploading...")
	if err := ftp.Mkd(ctx.path); err != nil {
		return err
	}

	if err := ftp.Cwd(ctx.path); err != nil {
		return err
	}

	err = ftp.Upload(archivePath)
	if err != nil {
		return err
	}

	logger.Info("Store successed")
	return nil
}
