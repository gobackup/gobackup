package compressor

import (
	"os/exec"
	"path/filepath"

	"github.com/gobackup/gobackup/helper"
)

type Tar struct {
	Base
}

func (tar *Tar) perform() (archivePath string, err error) {
	filePath := tar.archiveFilePath(tar.ext)

	var includes []string

	opts := tar.options()
	opts = append(opts, filePath)

	if tar.model.Archive != nil {
		includes = tar.model.Archive.GetStringSlice("includes")
		includes = cleanPaths(includes)

		if len(includes) >= 0 {

			excludes := tar.model.Archive.GetStringSlice("excludes")
			excludes = cleanPaths(excludes)

			for _, exclude := range excludes {
				opts = append(opts, "--exclude="+filepath.Clean(exclude))
			}
		}
	}

	opts = append(opts, tar.name)
	if includes != nil {
		opts = append(opts, includes...)
	}

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

func cleanPaths(paths []string) (results []string) {
	for _, p := range paths {
		results = append(results, filepath.Clean(p))
	}
	return
}
