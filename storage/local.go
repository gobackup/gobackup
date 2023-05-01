package storage

import (
	"fmt"
	"os"
	"path"
	"path/filepath"

	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
)

// Local storage
//
// type: local
// path: /data/backups
type Local struct {
	Base
	path string
}

func (s *Local) open() error {
	s.path = s.viper.GetString("path")
	return helper.MkdirP(s.path)
}

func (s *Local) close() {}

func (s *Local) upload(fileKey string) (err error) {
	logger := logger.Tag("Local")

	targetPath := path.Join(s.path, fileKey)
	targetDir := path.Dir(targetPath)
	if err := helper.MkdirP(targetDir); err != nil {
		return fmt.Errorf("mkdir %q: %w", targetDir, err)
	}

	_, err = helper.Exec("cp", "-a", s.archivePath, targetPath)
	if err != nil {
		return err
	}
	logger.Info("Store succeeded", targetPath)
	return nil
}

func (s *Local) delete(fileKey string) (err error) {
	targetPath := filepath.Join(s.path, fileKey)
	logger.Info("Deleting", targetPath)

	return os.Remove(targetPath)
}

// List all files
func (s *Local) list(parent string) ([]FileItem, error) {
	remotePath := filepath.Join(s.path, parent)
	items := []FileItem{}

	files, err := os.ReadDir(remotePath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() {
			info, err := file.Info()
			if err != nil {
				return nil, fmt.Errorf("get file info %q: %w", file.Name(), err)
			}

			items = append(items, FileItem{
				Filename:     file.Name(),
				Size:         info.Size(),
				LastModified: info.ModTime(),
			})
		}
	}

	return items, nil
}

func (s *Local) download(fileKey string) (string, error) {
	return "", fmt.Errorf("Local is not support download")
}
