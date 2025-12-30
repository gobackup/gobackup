package storage

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/studio-b12/gowebdav"

	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
)

// WebDAV storage
//
// # Install WebDAV Server on macOS
// https://github.com/hacdias/webdav/releases/tag/v4.2.0
//
//	 echo "users:\n  - username: admin\n    password: admin" > config.yml
//		./webdav --port 8080 -c config.yml
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
	logger := logger.Tag("WebDAV")
	logger.Info("-> Uploading...")

	var fileKeys []string
	if len(s.fileKeys) != 0 {
		// directory
		// 2022.12.04.07.09.47/2022.12.04.07.09.47.tar.xz-000
		fileKeys = s.fileKeys

		remotePath := filepath.Join(s.path, fileKey)
		remoteDir := filepath.Dir(remotePath)

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
		sourcePath := filepath.Join(filepath.Dir(s.archivePath), key)
		remotePath := filepath.Join(s.path, key)

		f, err := os.Open(sourcePath)
		if err != nil {
			return fmt.Errorf("failed to open file %q, %v", sourcePath, err)
		}
		defer f.Close()

		progress := helper.NewProgressBar(logger, f)
		if err := s.client.WriteStream(remotePath, progress.Reader, 0644); err != nil {
			return progress.Errorf("upload failed %v", err)
		}
		progress.Done(remotePath)
	}

	logger.Info("Store succeeded")
	return nil
}

func (s *WebDAV) delete(fileKey string) error {
	logger := logger.Tag("WebDAV")
	remotePath := path.Join(s.path, fileKey)
	logger.Info("-> remove", remotePath)
	return s.client.Remove(remotePath)
}

// List all files from storage
func (s *WebDAV) list(parent string) ([]FileItem, error) {
	remotePath := filepath.Join(s.path, parent)

	entries, err := s.client.ReadDir(remotePath)
	if err != nil {
		return nil, err
	}

	var items []FileItem
	for _, entry := range entries {
		if !entry.IsDir() {
			items = append(items, FileItem{
				Filename:     entry.Name(),
				Size:         entry.Size(),
				LastModified: entry.ModTime(),
			})
		}
	}

	return items, nil
}

func (s *WebDAV) download(fileKey string) (string, error) {
	return "", fmt.Errorf("WebDAV not support download")
}
