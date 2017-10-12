package encryptor

import (
	"fmt"
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/helper"
)

// OpenSSL encryptor for use openssl aes-256-cbc
//
// - base64: false
// - salt: true
// - password:
type OpenSSL struct {
	salt     bool
	base64   bool
	password string
}

func (ctx *OpenSSL) perform(archivePath string, model config.ModelConfig) (encryptPath string, err error) {
	sslViper := model.EncryptWith.Viper
	sslViper.SetDefault("salt", true)
	sslViper.SetDefault("base64", false)

	ctx.salt = sslViper.GetBool("salt")
	ctx.base64 = sslViper.GetBool("base64")
	ctx.password = sslViper.GetString("password")

	if len(ctx.password) == 0 {
		err = fmt.Errorf("password option is required")
		return
	}

	encryptPath = archivePath + ".enc"

	opts := ctx.options()
	opts = append(opts, "-in", archivePath, "-out", encryptPath)
	_, err = helper.Exec("openssl", opts...)
	return
}

func (ctx *OpenSSL) options() (opts []string) {
	opts = append(opts, "aes-256-cbc")
	if ctx.base64 {
		opts = append(opts, "-base64")
	}
	if ctx.salt {
		opts = append(opts, "-salt")
	}
	opts = append(opts, `-k`, ctx.password)
	return opts
}
