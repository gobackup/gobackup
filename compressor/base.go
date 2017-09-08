package compressor

import (
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/logger"
)

// Base compressor
type Base interface {
	perform() error
}

// Run compressor
func Run() error {
	logger.Info("------------- Compressor --------------")
	var ctx Base
	switch config.CompressWith {
	case "tgz":
		ctx = &Tgz{}
	default:
		ctx = &Tgz{}
	}

	err := ctx.perform()
	if err != nil {
		return err
	}
	logger.Info("----------- End Compressor ------------\n")

	return nil
}
