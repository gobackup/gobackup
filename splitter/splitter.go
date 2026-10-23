package splitter

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gobackup/gobackup/config"
	"github.com/gobackup/gobackup/helper"
	"github.com/gobackup/gobackup/hook"
	"github.com/gobackup/gobackup/logger"
	"github.com/spf13/viper"
)

// Run splitter
func Run(archivePath string, model config.ModelConfig) (archiveDirPath string, err error) {
	if model.Splitter == nil {
		return archivePath, nil
	}

	ctx := hook.Context{Model: model.Name, Stage: hook.StageSplit}
	startTime := time.Now()
	if err = hook.RunBefore(model.SplitterHooks, ctx); err == nil {
		archiveDirPath, err = run(archivePath, model)
	}
	hook.RunAfter(model.SplitterHooks, ctx, hook.Result{Err: err, Duration: time.Since(startTime)})

	if err != nil {
		return "", err
	}
	return archiveDirPath, nil
}

func run(archivePath string, model config.ModelConfig) (archiveDirPath string, err error) {
	logger := logger.Tag("Splitter")
	splitter := model.Splitter

	logger.Info("Split to chunks")

	splitter.SetDefault("suffix_length", 3)
	splitter.SetDefault("numeric_suffixes", true)
	if len(splitter.GetString("chunk_size")) == 0 {
		err = fmt.Errorf("chunk_size option is required")
		return
	}

	ext := model.Viper.GetString("Ext")
	// /tmp/gobackup3755903383/1670167448676759530/2022.12.04.07.24.08
	archiveDirPath = strings.TrimSuffix(archivePath, ext)
	if err = helper.MkdirP(archiveDirPath); err != nil {
		return
	}
	// /tmp/gobackup3755903383/1670167448676759530/2022.12.04.07.24.08/2022.12.04.07.24.08.tar.xz-
	splitSuffix := filepath.Join(archiveDirPath, filepath.Base(archivePath)+"-")

	opts := options(splitter)
	opts = append(opts, archivePath, splitSuffix)
	_, err = helper.Exec("split", opts...)
	if err != nil {
		return
	}

	logger.Info("Split done")

	err = os.Remove(archivePath)
	if err != nil {
		return
	}

	return
}

func options(splitter *viper.Viper) (opts []string) {
	bytes := splitter.GetString("chunk_size")
	opts = append(opts, "-b", bytes)
	suffixLength := splitter.GetInt("suffix_length")
	opts = append(opts, "-a", strconv.Itoa(suffixLength))
	if splitter.GetBool("numeric_suffixes") {
		opts = append(opts, "--numeric-suffixes")
	}

	return
}
