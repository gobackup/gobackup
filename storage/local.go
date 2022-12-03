package storage

import (
	"path"

	"github.com/huacnlee/gobackup/helper"
	"github.com/huacnlee/gobackup/logger"
)

// Local storage
//
// type: local
// path: /data/backups
type Local struct {
	Base
	destPath string
}

func (s *Local) open() (err error) {
	s.destPath = s.viper.GetString("path")
	helper.MkdirP(s.destPath)
	return
}

func (s *Local) close() {}

func (s *Local) upload(fileKey string) (err error) {
	logger := logger.Tag("Local")

	_, err = helper.Exec("cp", s.archivePath, s.destPath)
	if err != nil {
		return err
	}
	logger.Info("Store succeeded", s.destPath)
	return nil
}

func (s *Local) delete(fileKey string) (err error) {
	_, err = helper.Exec("rm", path.Join(s.destPath, fileKey))
	return
}
