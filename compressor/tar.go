package compressor

import (
	"github.com/huacnlee/gobackup/helper"
)

// Tar noop compressor
//
// type: tar (store only)
type Tar struct {
	Base
}

func (ctx *Tar) perform() (archivePath string, err error) {
	filePath := ctx.archiveFilePath(".tar")

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

func (ctx *Tar) options() (opts []string) {
	if helper.IsGnuTar {
		opts = append(opts, "--ignore-failed-read")
	}
	opts = append(opts, "-cf")

	return
}
