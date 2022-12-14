package encryptor

import (
	"fmt"

	"github.com/gobackup/gobackup/helper"
)

// OpenSSL encryptor for use openssl aes-256-cbc
//
// - base64: false
// - salt: true
// - password:
// - pbkdf2: true
// - iter: 10000
type OpenSSL struct {
	Base
	salt     bool
	base64   bool
	iter     int
	pbkdf2   bool
	password string
}

func (enc *OpenSSL) perform() (encryptPath string, err error) {
	sslViper := enc.viper
	sslViper.SetDefault("salt", true)
	sslViper.SetDefault("base64", false)
	sslViper.SetDefault("iter", 10000)
	sslViper.SetDefault("pbkdf2", true)

	enc.salt = sslViper.GetBool("salt")
	enc.base64 = sslViper.GetBool("base64")
	enc.password = sslViper.GetString("password")
	enc.pbkdf2 = sslViper.GetBool("pbkdf2")
	enc.password = sslViper.GetString("password")

	if len(enc.password) == 0 {
		err = fmt.Errorf("password option is required")
		return
	}

	encryptPath = enc.archivePath + ".enc"

	opts := enc.options()
	opts = append(opts, "-in", enc.archivePath, "-out", encryptPath)
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
	if enc.iter > 0 {
		opts = append(opts, "-iter", fmt.Sprintf("%d", enc.iter))
	}
	if enc.pbkdf2 {
		opts = append(opts, "-pbkdf2")
	}
	opts = append(opts, `-k`, enc.password)
	return opts
}
