package storage

import (
	"os"
	"path"

	"github.com/huacnlee/gobackup/helper"

	// "crypto/tls"
	"time"

	"github.com/huacnlee/gobackup/logger"
	"github.com/secsy/goftp"
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

	client *goftp.Client
}

func (ctx *FTP) open() (err error) {
	ctx.viper.SetDefault("port", "21")
	ctx.viper.SetDefault("timeout", 300)

	ctx.host = helper.CleanHost(ctx.viper.GetString("host"))
	ctx.port = ctx.viper.GetString("port")
	ctx.path = ctx.viper.GetString("path")
	ctx.username = ctx.viper.GetString("username")
	ctx.password = ctx.viper.GetString("password")

	ftpConfig := goftp.Config{
		User:     ctx.viper.GetString("username"),
		Password: ctx.viper.GetString("password"),
		Timeout:  ctx.viper.GetDuration("timeout") * time.Second,
	}
	ctx.client, err = goftp.DialConfig(ftpConfig, ctx.host+":"+ctx.port)
	if err != nil {
		return err
	}
	return
}

func (ctx *FTP) close() {
	ctx.client.Close()
}

func (ctx *FTP) upload(fileKey string) (err error) {
	logger.Info("-> Uploading...")
	_, err = ctx.client.Stat(ctx.path)
	if os.IsNotExist(err) {
		if _, err := ctx.client.Mkdir(ctx.path); err != nil {
			return err
		}
	}

	file, err := os.Open(ctx.archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	remotePath := path.Join(ctx.path, fileKey)
	err = ctx.client.Store(remotePath, file)
	if err != nil {
		return err
	}

	logger.Info("Store successed")
	return nil
}

func (ctx *FTP) delete(fileKey string) (err error) {
	remotePath := path.Join(ctx.path, fileKey)
	err = ctx.client.Delete(remotePath)
	return
}
