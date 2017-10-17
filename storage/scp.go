package storage

import (
	"github.com/huacnlee/gobackup/helper"
	"golang.org/x/crypto/ssh"
	"os"
	"path"
	"time"
	// "crypto/tls"
	"github.com/bramvdbogaerde/go-scp"
	"github.com/bramvdbogaerde/go-scp/auth"
	"github.com/huacnlee/gobackup/logger"
)

// SCP storage
//
// type: scp
// host: 192.168.1.2
// port: 22
// username: root
// password:
// timeout: 300
// private_key: ~/.ssh/id_rsa
type SCP struct {
	Base
	path       string
	host       string
	port       string
	privateKey string
	username   string
	password   string
}

func (ctx *SCP) perform() error {
	ctx.viper.SetDefault("port", "22")
	ctx.viper.SetDefault("timeout", 300)
	ctx.viper.SetDefault("private_key", "~/.ssh/id_rsa")

	ctx.host = ctx.viper.GetString("host")
	ctx.port = ctx.viper.GetString("port")
	ctx.path = ctx.viper.GetString("path")
	ctx.username = ctx.viper.GetString("username")
	ctx.password = ctx.viper.GetString("password")
	ctx.privateKey = helper.ExplandHome(ctx.viper.GetString("private_key"))

	logger.Info("PrivateKey", ctx.privateKey)

	clientConfig, _ := auth.PrivateKey(
		ctx.username,
		ctx.privateKey,
		ssh.InsecureIgnoreHostKey(),
	)
	clientConfig.Timeout = ctx.viper.GetDuration("timeout") * time.Second
	if len(ctx.password) > 0 {
		clientConfig.Auth = append(clientConfig.Auth, ssh.Password(ctx.password))
	}

	client := scp.NewClient(ctx.host+":"+ctx.port, &clientConfig)

	logger.Info("-> Connecting...")
	err := client.Connect()
	if err != nil {
		return err
	}
	defer client.Session.Close()

	file, err := os.Open(ctx.archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	remotePath := path.Join(ctx.path, ctx.fileKey)

	logger.Info("-> scp", remotePath)
	client.CopyFromFile(*file, remotePath, "0655")

	logger.Info("Store successed")
	return nil
}
