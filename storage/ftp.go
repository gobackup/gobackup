package storage

import (
	"os"
	"path"
	"path/filepath"
	// "crypto/tls"
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/logger"
	"github.com/secsy/goftp"
	"time"
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

	ftpViper := model.StoreWith.Viper

	ftpViper.SetDefault("port", "21")
	ftpViper.SetDefault("timeout", 300)

	ctx.host = ftpViper.GetString("host")
	ctx.port = ftpViper.GetString("port")
	ctx.path = ftpViper.GetString("path")
	ctx.username = ftpViper.GetString("username")
	ctx.password = ftpViper.GetString("password")

	ftpConfig := goftp.Config{
		User:     ftpViper.GetString("username"),
		Password: ftpViper.GetString("password"),
		Timeout:  ftpViper.GetDuration("timeout") * time.Second,
	}

	ftp, err := goftp.DialConfig(ftpConfig, ftpViper.GetString("host")+":"+ftpViper.GetString("port"))
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
	defer file.Close()

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
