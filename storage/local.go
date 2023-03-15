package storage

import (
	"fmt"
	"io/ioutil"
	"os"
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

	_, err = helper.Exec("cp", "-a", s.archivePath, s.path)
	if err != nil {
		return err
	}
	logger.Info("Store succeeded", filepath.Join(s.path, filepath.Base(s.archivePath)))
	return nil
}

func (s *Local) delete(fileKey string) (err error) {
	return os.Remove(filepath.Join(s.path, fileKey))
}

// List all files
func (s *Local) list(parent string) ([]FileItem, error) {
	remotePath := filepath.Join(s.path, parent)
	var items = []FileItem{}

	files, err := ioutil.ReadDir(remotePath)
	if err != nil {
		return nil, err
	}

	for _, file := range files {
		if !file.IsDir() {
			items = append(items, FileItem{
				Filename:     file.Name(),
				Size:         file.Size(),
				LastModified: file.ModTime(),
			})
		}
	}

	return items, nil
}

func (s *Local) download(fileKey string) (string, error) {
	return "", fmt.Errorf("Local is not support download")
}
