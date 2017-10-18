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

	hosts     string
	ftpConfig goftp.Config
}

func (ctx *FTP) init() (err error) {
	ctx.viper.SetDefault("port", "21")
	ctx.viper.SetDefault("timeout", 300)

	ctx.host = ctx.viper.GetString("host")
	ctx.port = ctx.viper.GetString("port")
	ctx.path = ctx.viper.GetString("path")
	ctx.username = ctx.viper.GetString("username")
	ctx.password = ctx.viper.GetString("password")

	ctx.ftpConfig = goftp.Config{
		User:     ctx.viper.GetString("username"),
		Password: ctx.viper.GetString("password"),
		Timeout:  ctx.viper.GetDuration("timeout") * time.Second,
	}
	ctx.hosts = ctx.host + ":" + ctx.port
	return
}

func (ctx *FTP) upload(fileKey string) (err error) {
	client, err := goftp.DialConfig(ctx.ftpConfig, ctx.hosts)
	if err != nil {
		return err
	}
	defer client.Close()

	logger.Info("-> Uploading...")
	_, err = client.Stat(ctx.path)
	if os.IsNotExist(err) {
		if _, err := client.Mkdir(ctx.path); err != nil {
			return err
		}
	}

	file, err := os.Open(ctx.archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	remotePath := path.Join(ctx.path, fileKey)
	err = client.Store(remotePath, file)
	if err != nil {
		return err
	}

	logger.Info("Store successed")
	return nil
}

func (ctx *FTP) delete(fileKey string) (err error) {
	client, err := goftp.DialConfig(ctx.ftpConfig, ctx.hosts)
	if err != nil {
		return err
	}
	defer client.Close()

	remotePath := path.Join(ctx.path, fileKey)
	err = client.Delete(remotePath)
	return
}
