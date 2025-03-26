package model

import (
	"fmt"
	"os"

	"github.com/spf13/viper"

	"github.com/KurosawaAngel/gobackup/archive"
	"github.com/KurosawaAngel/gobackup/compressor"
	"github.com/KurosawaAngel/gobackup/config"
	"github.com/KurosawaAngel/gobackup/database"
	"github.com/KurosawaAngel/gobackup/encryptor"
	"github.com/KurosawaAngel/gobackup/helper"
	"github.com/KurosawaAngel/gobackup/logger"
	"github.com/KurosawaAngel/gobackup/notifier"
	"github.com/KurosawaAngel/gobackup/splitter"
	"github.com/KurosawaAngel/gobackup/storage"
)

// Model class
type Model struct {
	Config config.ModelConfig
}

// Perform model
func (m Model) Perform() (err error) {
	logger := logger.Tag(fmt.Sprintf("Model: %s", m.Config.Name))

	m.before()

	defer func() {
		if err != nil {
			logger.Error(err)
			notifier.Failure(m.Config, err.Error())
		} else {
			notifier.Success(m.Config)
		}
	}()

	logger.Info("WorkDir:", m.Config.DumpPath)

	defer func() {
		if r := recover(); r != nil {
			m.after()
		}

		m.after()
	}()

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
	archivePath, err := compressor.Run(m.Config)
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

func (m Model) before() {
	// Execute before_script
	if len(m.Config.BeforeScript) > 0 {
		logger.Info("Executing before_script...")
		_, err := helper.ExecWithStdio(m.Config.BeforeScript, true)
		if err != nil {
			logger.Error(err)
		}
	}
}

// Cleanup model temp files
func (m Model) after() {
	logger := logger.Tag("Model")

	tempDir := m.Config.TempPath
	if viper.GetBool("useTempWorkDir") {
		tempDir = viper.GetString("workdir")
	}
	logger.Infof("Cleanup temp: %s/", tempDir)
	if err := os.RemoveAll(tempDir); err != nil {
		logger.Errorf("Cleanup temp dir %s error: %v", tempDir, err)
	}

	// Execute after_script
	if len(m.Config.AfterScript) > 0 {
		logger.Info("Executing after_script...")
		_, err := helper.ExecWithStdio(m.Config.AfterScript, true)
		if err != nil {
			logger.Error(err)
		}
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
