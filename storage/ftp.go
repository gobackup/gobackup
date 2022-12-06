package storage

import (
	"fmt"
	"net/textproto"
	"os"
	"path"
	"path/filepath"
	"strings"
	"time"

	"github.com/jlaffaye/ftp"

	"github.com/huacnlee/gobackup/helper"
	"github.com/huacnlee/gobackup/logger"
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

	client *ftp.ServerConn
}

func (s *FTP) open() error {
	s.viper.SetDefault("port", "21")
	s.viper.SetDefault("timeout", 300)
	s.viper.SetDefault("path", "/")

	s.host = helper.CleanHost(s.viper.GetString("host"))
	s.port = s.viper.GetString("port")
	s.path = s.viper.GetString("path")
	s.username = s.viper.GetString("username")
	s.password = s.viper.GetString("password")

	if len(s.host) == 0 || len(s.username) == 0 || len(s.password) == 0 {
		return fmt.Errorf("FTP host, username or password is empty")
	}

	timeout := s.viper.GetDuration("timeout") * time.Second
	client, err := ftp.Dial(s.host+":"+s.port, ftp.DialWithTimeout(timeout))
	if err != nil {
		return err
	}

	if err := client.Login(s.username, s.password); err != nil {
		return err
	}

	s.client = client

	return s.mkdir(s.path)
}

func (s *FTP) close() {
	s.client.Quit()
}

func (s *FTP) mkdir(rpath string) error {
	logger := logger.Tag("FTP")
	_, err := s.client.GetEntry(rpath)
	logger.Debugf("GetEntry %s: %v", rpath, err)
	if err != nil {
		if err.(*textproto.Error).Msg == "Can't check for file existence" {
			if err := s.client.MakeDir(rpath); err != nil {
				return err
			} else {
				return nil
			}
		} else {
			return err
		}
	}
	return nil
}

func (s *FTP) upload(fileKey string) error {
	logger := logger.Tag("FTP")
	logger.Info("-> Uploading...")

	var fileKeys []string
	if len(s.fileKeys) != 0 {
		// directory
		// 2022.12.04.07.09.47/2022.12.04.07.09.47.tar.xz-000
		fileKeys = s.fileKeys
		remoteDir := filepath.Join(s.path, filepath.Base(s.archivePath))
		// mkdir
		if err := s.mkdir(remoteDir); err != nil {
			return err
		}
	} else {
		// file
		// 2022.12.04.07.09.25.tar.xz
		fileKeys = append(fileKeys, fileKey)
	}

	for _, key := range fileKeys {
		filePath := filepath.Join(filepath.Dir(s.archivePath), key)
		f, err := os.Open(filePath)
		if err != nil {
			return fmt.Errorf("failed to open file %q, %v", filePath, err)
		}
		defer f.Close()

		remotePath := filepath.Join(s.path, key)
		if err := s.client.Stor(remotePath, f); err != nil {
			return err
		}

		logger.Infof("Store %s succeeded", remotePath)
	}

	logger.Info("Store succeeded")
	return nil
}

func (s *FTP) delete(fileKey string) error {
	logger := logger.Tag("FTP")
	remotePath := path.Join(s.path, fileKey)
	logger.Info("-> remove", remotePath)
	if !strings.HasSuffix(fileKey, "/") {
		// file
		return s.client.Delete(remotePath)
	} else {
		// directory
		return s.client.RemoveDir(remotePath)
	}

	return nil
}
