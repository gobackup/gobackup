package storage

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/studio-b12/gowebdav"

	"github.com/gobackup/gobackup/logger"
)

// WebDAV storage
//
// type: webdav
// root: http://localhost:8080
// username:
// password:
// path: backups
type WebDAV struct {
	Base
	path     string
	root     string
	username string
	password string
	client   *gowebdav.Client
	logger   *logger.Logger
}

func (s *WebDAV) open() error {
	s.root = s.viper.GetString("root")
	s.path = s.viper.GetString("path")
	s.username = s.viper.GetString("username")
	s.password = s.viper.GetString("password")

	if len(s.root) == 0 {
		return fmt.Errorf("WebDAV root is empty")
	}

	client := gowebdav.NewClient(s.root, s.username, s.password)
	if err := client.Connect(); err != nil {
		return err
	}

	s.client = client

	return s.client.MkdirAll(s.path, 0644)
}

func (s *WebDAV) close() {}

func (s *WebDAV) upload(fileKey string) error {
	logger := s.logger
	logger.Info("-> Uploading...")

	var fileKeys []string
	if len(s.fileKeys) != 0 {
		// directory
		// 2022.12.04.07.09.47/2022.12.04.07.09.47.tar.xz-000
		fileKeys = s.fileKeys
		remoteDir := filepath.Join(s.path, filepath.Base(s.archivePath))
		// mkdir
		if err := s.client.MkdirAll(remoteDir, 0644); err != nil {
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
		if err := s.client.WriteStream(remotePath, f, 0644); err != nil {
			return err
		}

		logger.Infof("Store %s succeeded", remotePath)
	}

	logger.Info("Store succeeded")
	return nil
}

func (s *WebDAV) delete(fileKey string) error {
	logger := s.logger
	remotePath := path.Join(s.path, fileKey)
	logger.Info("-> remove", remotePath)
	return s.client.Remove(remotePath)
}
