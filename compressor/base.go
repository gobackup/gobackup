package compressor

import (
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/logger"
)

// Base compressor
type Base interface {
	perform(model config.ModelConfig) (resultPath *string, err error)
}

// Run compressor
func Run(model config.ModelConfig) (resultPath *string, err error) {
	logger.Info("------------- Compressor --------------")
	var ctx Base
	switch model.CompressWith.Type {
	case "tgz":
		ctx = &Tgz{}
	default:
		ctx = &Tgz{}
	}

	resultPath, err = ctx.perform(model)
	if err != nil {
		return
	}
	logger.Info("----------- End Compressor ------------\n")

	return
}
