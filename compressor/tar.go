package compressor

import (
	"os/exec"

	"github.com/huacnlee/gobackup/helper"
)

type Tar struct {
	Base
}

func (ctx *Tar) perform() (archivePath string, err error) {
	filePath := ctx.archiveFilePath(ctx.ext)

	opts := ctx.options()
	opts = append(opts, filePath)
	opts = append(opts, ctx.name)
	archivePath = filePath

	_, err = helper.Exec("tar", opts...)

	return
}

func (ctx *Tar) options() (opts []string) {
	if helper.IsGnuTar {
		opts = append(opts, "--ignore-failed-read")
	}

	if ctx.ext == ".tar.gz" {
		path, err := exec.LookPath("pigz")
		if err == nil {
			opts = append(opts, "--use-compress-program", path)
		} else {
			opts = append(opts, "-a")
		}
	} else {
		opts = append(opts, "-a")
	}
	opts = append(opts, "-cf")

	return
}
