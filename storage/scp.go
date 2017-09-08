package storage

import (
	"golang.org/x/crypto/ssh"
	"os"
	"path"
	"path/filepath"
	"time"
	// "crypto/tls"
	"github.com/bramvdbogaerde/go-scp"
	"github.com/bramvdbogaerde/go-scp/auth"
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/logger"
)

// SCP storage
type SCP struct {
	path     string
	host     string
	port     string
	username string
	password string
}

func (ctx *SCP) perform(model config.ModelConfig, archivePath string) error {
	logger.Info("=> storage | SCP")

	scpViper := model.StoreWith.Viper

	scpViper.SetDefault("port", "22")
	scpViper.SetDefault("timeout", 300)

	ctx.host = scpViper.GetString("host")
	ctx.port = scpViper.GetString("port")
	ctx.path = scpViper.GetString("path")
	ctx.username = scpViper.GetString("username")
	ctx.password = scpViper.GetString("password")

	clientConfig, _ := auth.PrivateKey(ctx.username, scpViper.GetString("private_key"))
	clientConfig.Timeout = scpViper.GetDuration("timeout") * time.Second
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

	file, err := os.Open(archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	filename := filepath.Base(archivePath)
	remotePath := path.Join(ctx.path, filename)

	logger.Info("-> scp", remotePath)
	client.CopyFile(file, remotePath, "0655")

	logger.Info("Store successed")
	return nil
}
