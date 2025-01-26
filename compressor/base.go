package compressor

import (
	"fmt"
	"github.com/gobackup/gobackup/helper"
	"os"
	"path/filepath"
	"time"

	"github.com/gobackup/gobackup/config"
	"github.com/gobackup/gobackup/logger"
	"github.com/spf13/viper"
)

// Base compressor
type Base struct {
	name            string
	ext             string
	parallelProgram string
	model           config.ModelConfig
	viper           *viper.Viper
}

// Compressor
type Compressor interface {
	perform() (archivePath string, err error)
}

func (c *Base) archiveFilePath(ext string) string {
	return filepath.Join(c.model.TempPath, time.Now().Format("2006.01.02.15.04.05")+ext)
}

func newBase(model config.ModelConfig) (base Base) {
	base = Base{
		name:  model.Name,
		model: model,
		viper: model.CompressWith.Viper,
	}
	return
}

// Run compressor, return archive path
func Run(model config.ModelConfig) (string, error) {
	logger := logger.Tag("Compressor")

	base := newBase(model)

	var c Compressor
	var ext, parallelProgram string
	switch model.CompressWith.Type {
	case "gz", "tgz", "taz", "tar.gz":
		ext = ".tar.gz"
		parallelProgram = "pigz"
	case "Z", "taZ", "tar.Z":
		ext = ".tar.Z"
	case "bz2", "tbz", "tbz2", "tar.bz2":
		ext = ".tar.bz2"
		parallelProgram = "pbzip2"
	case "lz", "tar.lz":
		ext = ".tar.lz"
	case "lzma", "tlz", "tar.lzma":
		ext = ".tar.lzma"
	case "lzo", "tar.lzo":
		ext = ".tar.lzo"
	case "xz", "txz", "tar.xz":
		ext = ".tar.xz"
		parallelProgram = "pixz"
	case "zst", "tzst", "tar.zst":
		ext = ".tar.zst"
	case "tar":
		ext = ".tar"
	case "":
		ext = ".tar"
		model.CompressWith.Type = "tar"
	default:
		return "", fmt.Errorf("Unsupported compress type: %s", model.CompressWith.Type)
	}

	// save Extension
	model.Viper.Set("Ext", ext)

	base.ext = ext
	base.parallelProgram = parallelProgram
	c = &Tar{Base: base}

	logger.Info("=> Compress | " + model.CompressWith.Type)

	if err := helper.MkdirP(model.DumpPath); err != nil {
		logger.Errorf("Failed to mkdir dump path %s: %v", model.DumpPath, err)
		return "", err
	}

	// set workdir
	if err := os.Chdir(filepath.Join(model.DumpPath, "../")); err != nil {
		return "", fmt.Errorf("chdir to dump path: %s: %w", model.DumpPath, err)
	}

	archivePath, err := c.perform()
	if err != nil {
		return "", err
	}
	logger.Info("->", archivePath)

	return archivePath, nil
}
