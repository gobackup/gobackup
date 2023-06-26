package archive

import (
	"fmt"
	"path"
	"path/filepath"

	"github.com/gobackup/gobackup/config"
	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/logger"
)

// Run archive
func Run(model config.ModelConfig) error {
	logger := logger.Tag("Archive")

	if model.Archive == nil {
		return nil
	}

	if err := helper.MkdirP(model.DumpPath); err != nil {
		logger.Errorf("Failed to mkdir dump path %s: %v", model.DumpPath, err)
		return err
	}

	includes := model.Archive.GetStringSlice("includes")
	includes = cleanPaths(includes)

	excludes := model.Archive.GetStringSlice("excludes")
	excludes = cleanPaths(excludes)

	if len(includes) == 0 {
		return fmt.Errorf("archive.includes have no config")
	}
	logger.Info("=> includes", len(includes), "rules")

	opts := options(model.DumpPath, excludes, includes)

	_, err := helper.Exec("tar", opts...)
	return err
}

func options(dumpPath string, excludes, includes []string) (opts []string) {
	tarPath := path.Join(dumpPath, "archive.tar")
	if helper.IsGnuTar {
		opts = append(opts, "--ignore-failed-read")
	}
	opts = append(opts, "-cPf", tarPath)

	for _, exclude := range excludes {
		opts = append(opts, "--exclude="+filepath.Clean(exclude))
	}

	opts = append(opts, includes...)

	return opts
}

func cleanPaths(paths []string) (results []string) {
	for _, p := range paths {
		results = append(results, filepath.Clean(p))
	}
	return
}
