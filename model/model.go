package model

import (
	"fmt"
	"os"
	"runtime/debug"
	"time"

	"github.com/spf13/viper"

	"github.com/gobackup/gobackup/archive"
	"github.com/gobackup/gobackup/compressor"
	"github.com/gobackup/gobackup/config"
	"github.com/gobackup/gobackup/database"
	"github.com/gobackup/gobackup/encryptor"
	"github.com/gobackup/gobackup/hook"
	"github.com/gobackup/gobackup/logger"
	"github.com/gobackup/gobackup/metrics"
	"github.com/gobackup/gobackup/notifier"
	"github.com/gobackup/gobackup/splitter"
	"github.com/gobackup/gobackup/storage"
)

// Model class
type Model struct {
	Config config.ModelConfig
}

// Perform model
func (m Model) Perform() (err error) {
	logger := logger.Tag(fmt.Sprintf("Model: %s", m.Config.Name))
	startTime := time.Now()
	var archivePath string

	defer func() {
		duration := time.Since(startTime).Seconds()
		metrics.DuractionSeconds.WithLabelValues(m.Config.Name).Observe(duration)

		if err != nil {
			logger.Error(err)
			notifier.Failure(m.Config, err.Error())
			metrics.TotalAttempts.WithLabelValues(m.Config.Name, "failure").Inc()
			metrics.LastTimestamp.WithLabelValues(m.Config.Name, "failure").Set(float64(time.Now().Unix()))
		} else {
			notifier.Success(m.Config)
			metrics.TotalAttempts.WithLabelValues(m.Config.Name, "success").Inc()
			metrics.LastTimestamp.WithLabelValues(m.Config.Name, "success").Set(float64(time.Now().Unix()))

			// Record backup file size if available
			if archivePath != "" {
				if fi, statErr := os.Stat(archivePath); statErr == nil {
					metrics.FileSizes.WithLabelValues(m.Config.Name).Set(float64(fi.Size()))
				}
			}
		}
	}()

	logger.Info("WorkDir:", m.Config.DumpPath)

	hooks := m.hooks()

	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("panic: %v\n%s", r, debug.Stack())
			err = fmt.Errorf("panic: %v", r)
		}

		m.after(hooks, hook.Result{Err: err, Duration: time.Since(startTime)})
	}()

	if err = m.before(hooks); err != nil {
		return
	}

	err = database.Run(m.Config)
	if err != nil {
		return
	}

	if m.Config.Archive != nil {
		err = archive.Run(m.Config)
		if err != nil {
			return
		}
	}

	// It always to use compressor, default use tar, even not enable compress.
	archivePath, err = compressor.Run(m.Config)
	if err != nil {
		return
	}

	archivePath, err = encryptor.Run(archivePath, m.Config)
	if err != nil {
		return
	}

	archivePath, err = splitter.Run(archivePath, m.Config)
	if err != nil {
		return
	}

	err = storage.Run(m.Config, archivePath)
	if err != nil {
		return
	}

	return nil
}

func (m Model) hookContext() hook.Context {
	return hook.Context{Model: m.Config.Name, Stage: hook.StageRun}
}

// hooks returns model run hooks with legacy before_script / after_script
// prepended as implicit hooks. Legacy semantics are kept: before_script
// failure is only logged, after_script always runs.
func (m Model) hooks() []config.HookConfig {
	hooks := []config.HookConfig{}
	if len(m.Config.BeforeScript) > 0 {
		hooks = append(hooks, config.HookConfig{
			Name:            "before_script",
			When:            config.HookBefore,
			Script:          m.Config.BeforeScript,
			ContinueOnError: true,
		})
	}
	if len(m.Config.AfterScript) > 0 {
		hooks = append(hooks, config.HookConfig{
			Name:   "after_script",
			When:   config.HookAfter,
			OnExit: config.HookOnExitAlways,
			Script: m.Config.AfterScript,
		})
	}
	return append(hooks, m.Config.Hooks...)
}

// Run before_script and before hooks
func (m Model) before(hooks []config.HookConfig) error {
	return hook.RunBefore(hooks, m.hookContext())
}

// Run after_script and after hooks, then cleanup model temp files
func (m Model) after(hooks []config.HookConfig, result hook.Result) {
	logger := logger.Tag("Model")

	hook.RunAfter(hooks, m.hookContext(), result)

	tempDir := m.Config.TempPath
	if viper.GetBool("useTempWorkDir") {
		tempDir = viper.GetString("workdir")
	}
	logger.Infof("Cleanup temp: %s/", tempDir)
	if err := os.RemoveAll(tempDir); err != nil {
		logger.Errorf("Cleanup temp dir %s error: %v", tempDir, err)
	}
}

// GetModelByName get model by name
func GetModelByName(name string) *Model {
	modelConfig := config.GetModelConfigByName(name)
	if modelConfig == nil {
		return nil
	}
	return &Model{
		Config: *modelConfig,
	}
}

// GetModels get models
func GetModels() (models []*Model) {
	for _, modelConfig := range config.Models {
		m := Model{
			Config: modelConfig,
		}
		models = append(models, &m)
	}
	return
}
