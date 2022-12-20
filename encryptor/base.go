package encryptor

import (
	"github.com/gobackup/gobackup/config"
	"github.com/gobackup/gobackup/logger"
	"github.com/spf13/viper"
)

// Base encryptor
type Base struct {
	model       config.ModelConfig
	viper       *viper.Viper
	archivePath string
}

// Encryptor interface
type Encryptor interface {
	perform() (encryptPath string, err error)
}

func newBase(archivePath string, model config.ModelConfig) (base *Base) {
	base = &Base{
		archivePath: archivePath,
		model:       model,
		viper:       model.EncryptWith.Viper,
	}
	return
}

// Run compressor
func Run(archivePath string, model config.ModelConfig) (encryptPath string, err error) {
	logger := logger.Tag("Encryptor")

	base := newBase(archivePath, model)
	var enc Encryptor
	switch model.EncryptWith.Type {
	case "openssl":
		enc = NewOpenSSL(base)
	default:
		encryptPath = archivePath
		return
	}

	logger.Info("encrypt | " + model.EncryptWith.Type)
	encryptPath, err = enc.perform()
	if err != nil {
		return
	}
	logger.Info("encrypted:", encryptPath)

	// save Extension
	model.Viper.Set("Ext", model.Viper.GetString("Ext")+".enc")

	return
}
