package compressor

import (
	"github.com/huacnlee/gobackup/helper"
)

// Tgz .tar.gz compressor
//
// type: tgz
type Tgz struct {
	Base
}

func (ctx *Tgz) perform() (archivePath string, err error) {
	filePath := ctx.archiveFilePath(".tar.gz")

	opts := ctx.options()
	opts = append(opts, filePath)
	opts = append(opts, ctx.name)

	_, err = helper.Exec("tar", opts...)
	if err == nil {
		archivePath = filePath
		return
	}
	return
}

func (ctx *Tgz) options() (opts []string) {
	if helper.IsGnuTar {
		opts = append(opts, "--ignore-failed-read")
	}
	opts = append(opts, "-zcf")

	return
}
