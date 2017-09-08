package main

import (
	"github.com/huacnlee/gobackup/logger"

	"github.com/huacnlee/gobackup/config"
)

func main() {
	logger.Info(config.Models)
	for _, modelConfig := range config.Models {
		model := Model{
			Config: modelConfig,
		}
		model.perform()
	}
}
