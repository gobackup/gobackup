package encryptor

import (
	"fmt"

	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
)

// OpenSSL encryptor for use openssl aes-256-cbc
//
// - base64: false
// - salt: true
// - password:
// - args:
type OpenSSL struct {
	Base
	salt        bool
	base64      bool
	password    string
	args        string
	encryptPath string
}

func NewOpenSSL(base *Base) *OpenSSL {
	base.viper.SetDefault("salt", true)
	base.viper.SetDefault("base64", false)
	base.viper.SetDefault("args", "")

	return &OpenSSL{
		Base:        *base,
		salt:        base.viper.GetBool("salt"),
		base64:      base.viper.GetBool("base64"),
		password:    base.viper.GetString("password"),
		args:        base.viper.GetString("args"),
		encryptPath: base.archivePath + ".enc",
	}
}

func (enc *OpenSSL) perform() (encryptPath string, err error) {
	if len(enc.password) == 0 {
		err = fmt.Errorf("password option is required")
		return
	}

	opts := enc.options()
	opts = append(opts, "-in", enc.archivePath, "-out", enc.encryptPath)
	logger.Infof("openssl %s", opts)
	_, err = helper.Exec("openssl", opts...)
	return
}

func (enc *OpenSSL) options() (opts []string) {
	opts = append(opts, "aes-256-cbc")
	if enc.base64 {
		opts = append(opts, "-base64")
	}
	if enc.salt {
		opts = append(opts, "-salt")
	}
	if len(enc.args) > 0 {
		opts = append(opts, enc.args)
	}

	opts = append(opts, `-k`, enc.password)
	return opts
}
