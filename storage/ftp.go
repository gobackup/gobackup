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

func (s *FTP) open() (err error) {
	s.viper.SetDefault("port", "21")
	s.viper.SetDefault("timeout", 300)

	s.host = helper.CleanHost(s.viper.GetString("host"))
	s.port = s.viper.GetString("port")
	s.path = s.viper.GetString("path")
	s.username = s.viper.GetString("username")
	s.password = s.viper.GetString("password")

	ftpConfig := goftp.Config{
		User:     s.viper.GetString("username"),
		Password: s.viper.GetString("password"),
		Timeout:  s.viper.GetDuration("timeout") * time.Second,
	}
	s.client, err = goftp.DialConfig(ftpConfig, s.host+":"+s.port)
	if err != nil {
		return err
	}
	return
}

func (s *FTP) close() {
	s.client.Close()
}

func (s *FTP) upload(fileKey string) (err error) {
	logger := logger.Tag("FTP")

	logger.Info("-> Uploading...")
	_, err = s.client.Stat(s.path)
	if os.IsNotExist(err) {
		if _, err := s.client.Mkdir(s.path); err != nil {
			return err
		}
	}

	file, err := os.Open(s.archivePath)
	if err != nil {
		return err
	}
	defer file.Close()

	remotePath := path.Join(s.path, fileKey)
	err = s.client.Store(remotePath, file)
	if err != nil {
		return err
	}

	logger.Info("Store succeeded")
	return nil
}

func (s *FTP) delete(fileKey string) (err error) {
	remotePath := path.Join(s.path, fileKey)
	err = s.client.Delete(remotePath)
	return
}
