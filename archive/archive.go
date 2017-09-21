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

	tarPath := path.Join(model.DumpPath, "archive.tar")

	includes := model.Archive.GetStringSlice("includes")
	includes = cleanPaths(includes)

	if len(includes) == 0 {
		return fmt.Errorf("archive.includes have no config")
	}
	logger.Info("=> includes", len(includes), "rules")

	cmd := "tar -cPf " + tarPath

	excludes := model.Archive.GetStringSlice("excludes")
	excludes = cleanPaths(excludes)

	for _, exclude := range excludes {
		cmd += " --exclude='" + filepath.Clean(exclude) + "'"
	}

	helper.Exec(cmd, includes...)

	return nil
}

func cleanPaths(paths []string) (results []string) {
	for _, p := range paths {
		results = append(results, filepath.Clean(p))
	}
	return
}
