package encryptor

import (
	"time"

	"github.com/gobackup/gobackup/config"
	"github.com/gobackup/gobackup/hook"
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
		if len(model.EncryptWith.Hooks) > 0 {
			logger.Warnf("encrypt_with.hooks are ignored because encrypt type %q is not supported", model.EncryptWith.Type)
		}
		encryptPath = archivePath
		return
	}

	ctx := hook.Context{Model: model.Name, Stage: hook.StageEncrypt}
	startTime := time.Now()
	if err = hook.RunBefore(model.EncryptWith.Hooks, ctx); err == nil {
		logger.Info("encrypt | " + model.EncryptWith.Type)
		encryptPath, err = enc.perform()
	}
	hook.RunAfter(model.EncryptWith.Hooks, ctx, hook.Result{Err: err, Duration: time.Since(startTime)})
	if err != nil {
		return "", err
	}
	logger.Info("encrypted:", encryptPath)

	// save Extension
	model.Viper.Set("Ext", model.Viper.GetString("Ext")+".enc")

	return
}
