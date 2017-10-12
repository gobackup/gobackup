package encryptor

import (
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/logger"
)

// Base encryptor
type Base interface {
	perform(archivePath string, model config.ModelConfig) (encryptPath string, err error)
}

// Run compressor
func Run(archivePath string, model config.ModelConfig) (encryptPath string, err error) {
	var ctx Base
	switch model.EncryptWith.Type {
	case "openssl":
		ctx = &OpenSSL{}
	default:
		encryptPath = archivePath
		return
	}

	logger.Info("------------ Encryptor -------------")
	logger.Info("=> Encrypt | " + model.EncryptWith.Type)
	encryptPath, err = ctx.perform(archivePath, model)
	if err != nil {
		return
	}
	logger.Info("->", encryptPath)
	logger.Info("------------ Encryptor -------------\n")

	return
}
