package encryptor

import (
	"fmt"
	"strings"

	"github.com/gobackup/gobackup/helper"
)

// OpenSSL encryptor for use openssl aes-256-cbc
//
// - chiper: aes-256-cbc
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
	chiper      string
	encryptPath string
}

func NewOpenSSL(base *Base) *OpenSSL {
	base.viper.SetDefault("salt", true)
	base.viper.SetDefault("base64", false)
	base.viper.SetDefault("args", "")
	base.viper.SetDefault("chiper", "aes-256-cbc")

	return &OpenSSL{
		Base:        *base,
		salt:        base.viper.GetBool("salt"),
		base64:      base.viper.GetBool("base64"),
		password:    base.viper.GetString("password"),
		args:        base.viper.GetString("args"),
		chiper:      base.viper.GetString("chiper"),
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
	_, err = helper.Exec("openssl", opts...)
	if err != nil {
		err = fmt.Errorf("OpenSSL encrypt failed: %s `openssl %s`", strings.TrimSpace(err.Error()), strings.Join(opts, " "))
		return "", err
	}
	return enc.encryptPath, nil
}

func (enc *OpenSSL) options() (opts []string) {
	opts = append(opts, enc.chiper)
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
