package compressor

import (
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/logger"
)

// Base compressor
type Base interface {
	perform() (resultPath *string, err error)
}

// Run compressor
func Run() (resultPath *string, err error) {
	logger.Info("------------- Compressor --------------")
	var ctx Base
	switch config.CompressWith.Type {
	case "tgz":
		ctx = &Tgz{}
	default:
		ctx = &Tgz{}
	}

	resultPath, err = ctx.perform()
	if err != nil {
		return
	}
	logger.Info("----------- End Compressor ------------\n")

	return
}
