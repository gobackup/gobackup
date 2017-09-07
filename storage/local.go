package storage

import (
	"github.com/huacnlee/gobackup/logger"
)

type Local struct {
}

func newLocal() *Local {
	return &Local{}
}

func (ctx *Local) Perform() error {
	logger.Info("Storage local")
	return nil
}
