package model

import (
	"os"

	"github.com/spf13/viper"

	"github.com/gobackup/gobackup/archive"
	"github.com/gobackup/gobackup/compressor"
	"github.com/gobackup/gobackup/config"
	"github.com/gobackup/gobackup/database"
	"github.com/gobackup/gobackup/encryptor"
	"github.com/gobackup/gobackup/logger"
	"github.com/gobackup/gobackup/splitter"
	"github.com/gobackup/gobackup/storage"
)

// Model class
type Model struct {
	Config config.ModelConfig
	Logger *logger.Logger
}

// Perform model
func (m Model) Perform() {
	logger := m.Logger
	logger.Info("WorkDir:", m.Config.DumpPath)

	defer func() {
		if r := recover(); r != nil {
			m.cleanup()
		}

		m.cleanup()
	}()

	err := database.Run(m.Config, *logger)
	if err != nil {
		logger.Error(err)
		return
	}

	if m.Config.Archive != nil {
		err = archive.Run(m.Config, *logger)
		if err != nil {
			logger.Error(err)
			return
		}
	}

	archivePath, err := compressor.Run(m.Config, *logger)
	if err != nil {
		logger.Error(err)
		return
	}

	archivePath, err = encryptor.Run(archivePath, m.Config, *logger)
	if err != nil {
		logger.Error(err)
		return
	}

	archivePath, err = splitter.Run(archivePath, m.Config, *logger)
	if err != nil {
		logger.Error(err)
		return
	}

	err = storage.Run(m.Config, archivePath, *logger)
	if err != nil {
		logger.Error(err)
		return
	}

}

// Cleanup model temp files
func (m Model) cleanup() {
	logger := m.Logger

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
		Logger: logger.NewTagLogger(name),
	}
}

// GetModels get models
func GetModels() (models []*Model) {
	for _, modelConfig := range config.Models {
		m := Model{
			Config: modelConfig,
			Logger: logger.NewTagLogger(modelConfig.Name),
		}
		models = append(models, &m)
	}
	return
}
