package compressor

import (
	"os/exec"

	"github.com/gobackup/gobackup/helper"
)

type Tar struct {
	Base
}

func (tar *Tar) perform() (archivePath string, err error) {
	filePath := tar.archiveFilePath(tar.ext)

	opts := tar.options()
	opts = append(opts, filePath)
	opts = append(opts, tar.name)
	archivePath = filePath

	_, err = helper.Exec("tar", opts...)

	return
}

func (tar *Tar) options() (opts []string) {
	if helper.IsGnuTar {
		opts = append(opts, "--ignore-failed-read")
	}

	var useCompressProgram bool
	if len(tar.parallelProgram) > 0 {
		if path, err := exec.LookPath(tar.parallelProgram); err == nil {
			useCompressProgram = true
			opts = append(opts, "--use-compress-program", path)
		}
	}
	if !useCompressProgram {
		opts = append(opts, "-a")
	}
	opts = append(opts, "-cf")

	return
}
