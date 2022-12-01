package model

import (
	"fmt"
	"os"

	"github.com/huacnlee/gobackup/archive"
	"github.com/huacnlee/gobackup/compressor"
	"github.com/huacnlee/gobackup/config"
	"github.com/huacnlee/gobackup/database"
	"github.com/huacnlee/gobackup/encryptor"
	"github.com/huacnlee/gobackup/logger"
	"github.com/huacnlee/gobackup/storage"
)

// Model class
type Model struct {
	Config config.ModelConfig
}

// Perform model
func (m Model) Perform() {
	logger := logger.Tag(fmt.Sprintf("Modal: %s", m.Config.Name))

	logger.Info("WorkDir:", m.Config.DumpPath)

	defer func() {
		if r := recover(); r != nil {
			m.cleanup()
		}

		m.cleanup()
	}()

	err := database.Run(m.Config)
	if err != nil {
		logger.Error(err)
		return
	}

	if m.Config.Archive != nil {
		err = archive.Run(m.Config)
		if err != nil {
			logger.Error(err)
			return
		}
	}

	archivePath, err := compressor.Run(m.Config)
	if err != nil {
		logger.Error(err)
		return
	}

	archivePath, err = encryptor.Run(archivePath, m.Config)
	if err != nil {
		logger.Error(err)
		return
	}

	err = storage.Run(m.Config, archivePath)
	if err != nil {
		logger.Error(err)
		return
	}

}

// Cleanup model temp files
func (m Model) cleanup() {
	logger := logger.Tag("Model")

	logger.Info("Cleanup temp: " + m.Config.TempPath + "/")
	err := os.RemoveAll(m.Config.TempPath)
	if err != nil {
		logger.Error("Cleanup temp dir "+m.Config.TempPath+" error:", err)
	}
}
