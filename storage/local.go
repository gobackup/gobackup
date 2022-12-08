package storage

import (
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
	destPath string
	logger   *logger.Logger
}

func (s *Local) open() error {
	s.destPath = s.viper.GetString("path")
	return helper.MkdirP(s.destPath)
}

func (s *Local) close() {}

func (s *Local) upload(fileKey string) (err error) {
	logger := s.logger

	_, err = helper.Exec("cp", "-a", s.archivePath, s.destPath)
	if err != nil {
		return err
	}
	logger.Info("Store succeeded", filepath.Join(s.destPath, filepath.Base(s.archivePath)))
	return nil
}

func (s *Local) delete(fileKey string) (err error) {
	return os.Remove(filepath.Join(s.destPath, fileKey))
}
