package storage

import (
	"os"
	"path"
	"time"

	"github.com/huacnlee/gobackup/helper"
	"golang.org/x/crypto/ssh"

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
	client     scp.Client
}

func (s *SCP) open() (err error) {
	s.viper.SetDefault("port", "22")
	s.viper.SetDefault("timeout", 300)
	s.viper.SetDefault("private_key", "~/.ssh/id_rsa")

	s.host = s.viper.GetString("host")
	s.port = s.viper.GetString("port")
	s.path = s.viper.GetString("path")
	s.username = s.viper.GetString("username")
	s.password = s.viper.GetString("password")
	s.privateKey = helper.ExplandHome(s.viper.GetString("private_key"))
	var clientConfig ssh.ClientConfig
	logger.Info("PrivateKey", s.privateKey)
	clientConfig, err = auth.PrivateKey(
		s.username,
		s.privateKey,
		ssh.InsecureIgnoreHostKey(),
	)
	if err != nil {
		logger.Warn(err)
		logger.Info("PrivateKey fail, Try User@Host with Password")
		clientConfig = ssh.ClientConfig{
			User:            s.username,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
	}
	clientConfig.Timeout = s.viper.GetDuration("timeout") * time.Second
	if len(s.password) > 0 {
		clientConfig.Auth = append(clientConfig.Auth, ssh.Password(s.password))
	}

	s.client = scp.NewClient(s.host+":"+s.port, &clientConfig)

	err = s.client.Connect()
	if err != nil {
		return err
	}
	defer s.client.Session.Close()
	s.client.Session.Run("mkdir -p " + s.path)
	return
}

func (s *SCP) close() {}

func (s *SCP) upload(fileKey string) (err error) {
	err = s.client.Connect()
	if err != nil {
		return err
	}
	defer s.client.Session.Close()

	file, err := os.Open(s.archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	remotePath := path.Join(s.path, fileKey)
	logger.Info("-> scp", remotePath)
	s.client.CopyFromFile(*file, remotePath, "0655")

	logger.Info("Store successed")
	return nil
}

func (s *SCP) delete(fileKey string) (err error) {
	err = s.client.Connect()
	if err != nil {
		return
	}
	defer s.client.Session.Close()

	remotePath := path.Join(s.path, fileKey)
	logger.Info("-> remove", remotePath)
	err = s.client.Session.Run("rm " + remotePath)
	return
}
