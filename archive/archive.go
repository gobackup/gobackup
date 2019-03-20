package archive

import (
	"fmt"
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/helper"
	"github.com/huacnlee/gobackup/logger"
	"path"
	"path/filepath"
)

// Run archive
func Run(model config.ModelConfig) (err error) {
	if model.Archive == nil {
		return nil
	}

	logger.Info("------------- Archives -------------")

	helper.MkdirP(model.DumpPath)

	includes := model.Archive.GetStringSlice("includes")
	includes = cleanPaths(includes)

	excludes := model.Archive.GetStringSlice("excludes")
	excludes = cleanPaths(excludes)

	if len(includes) == 0 {
		return fmt.Errorf("archive.includes have no config")
	}
	logger.Info("=> includes", len(includes), "rules")

	opts := options(model.DumpPath, excludes, includes)
	helper.Exec("tar", opts...)

	logger.Info("------------- Archives -------------\n")

	return nil
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
