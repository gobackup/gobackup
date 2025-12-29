package storage

import (
	"fmt"
	"io/ioutil"
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

	// Related path
	if !path.IsAbs(s.path) {
		s.path = path.Join(s.model.WorkDir, s.path)
	}

	targetPath := path.Join(s.path, fileKey)
	targetDir := path.Dir(targetPath)
	if err := helper.MkdirP(targetDir); err != nil {
		logger.Errorf("failed to mkdir %q, %v", targetDir, err)
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

// uploadState for local storage stores state in the backup path
// This is useful when the local path is on a persistent volume
func (s *Local) uploadState(key string, data []byte) error {
	statePath := path.Join(s.path, key)
	stateDir := path.Dir(statePath)

	if err := helper.MkdirP(stateDir); err != nil {
		return fmt.Errorf("failed to create state directory %q: %v", stateDir, err)
	}

	if err := os.WriteFile(statePath, data, 0660); err != nil {
		return fmt.Errorf("failed to write state file %q: %v", statePath, err)
	}

	return nil
}

// downloadState for local storage reads state from the backup path
func (s *Local) downloadState(key string) ([]byte, error) {
	statePath := path.Join(s.path, key)

	data, err := os.ReadFile(statePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file %q: %v", statePath, err)
	}

	return data, nil
}
