package compressor

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/logger"
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

// Context compressor
type Context interface {
	perform() (archivePath string, err error)
}

func (ctx *Base) archiveFilePath(ext string) string {
	return path.Join(ctx.model.TempPath, time.Now().Format("2006.01.02.15.04.05")+ext)
}

func newBase(model config.ModelConfig) (base Base) {
	base = Base{
		name:  model.Name,
		model: model,
		viper: model.CompressWith.Viper,
	}
	return
}

// Run compressor
func Run(model config.ModelConfig) (archivePath string, err error) {
	logger := logger.Tag("Compressor")

	base := newBase(model)

	var ctx Context
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
		err = fmt.Errorf("Unsupported compress type: %s", model.CompressWith.Type)
		return
	}

	base.ext = ext
	base.parallelProgram = parallelProgram
	ctx = &Tar{Base: base}

	logger.Info("=> Compress | " + model.CompressWith.Type)

	// set workdir
	os.Chdir(path.Join(model.DumpPath, "../"))
	archivePath, err = ctx.perform()
	if err != nil {
		return
	}
	logger.Info("->", archivePath)

	return
}
