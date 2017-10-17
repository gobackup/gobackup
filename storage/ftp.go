package storage

import (
	"os"
	"path"
	// "crypto/tls"
	"github.com/huacnlee/gobackup/logger"
	"github.com/secsy/goftp"
	"time"
)

// FTP storage
//
// type: ftp
// path: /backups
// host: ftp.your-host.com
// port: 21
// timeout: 30
// username:
// password:
type FTP struct {
	Base
	path     string
	host     string
	port     string
	username string
	password string
}

func (ctx *FTP) perform() error {
	ctx.viper.SetDefault("port", "21")
	ctx.viper.SetDefault("timeout", 300)

	ctx.host = ctx.viper.GetString("host")
	ctx.port = ctx.viper.GetString("port")
	ctx.path = ctx.viper.GetString("path")
	ctx.username = ctx.viper.GetString("username")
	ctx.password = ctx.viper.GetString("password")

	ftpConfig := goftp.Config{
		User:     ctx.viper.GetString("username"),
		Password: ctx.viper.GetString("password"),
		Timeout:  ctx.viper.GetDuration("timeout") * time.Second,
	}

	ftp, err := goftp.DialConfig(
		ftpConfig,
		ctx.viper.GetString("host")+":"+ctx.viper.GetString("port"),
	)
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

	file, err := os.Open(ctx.archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	remotePath := path.Join(ctx.path, ctx.fileKey)
	logger.Info("-> upload", remotePath)
	err = ftp.Store(remotePath, file)
	if err != nil {
		return err
	}

	logger.Info("Store successed")
	return nil
}
