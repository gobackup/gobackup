package scheduler

import (
	"fmt"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/gobackup/gobackup/config"
	"github.com/gobackup/gobackup/logger"
	"github.com/gobackup/gobackup/model"
)

var (
	mycron *gocron.Scheduler
)

// Start scheduler
func Start() error {
	logger := logger.Tag("Scheduler")

	mycron = gocron.NewScheduler(time.Local)

	for _, modelConfig := range config.Models {
		if !modelConfig.Schedule.Enabled {
			continue
		}

		logger.Info(fmt.Sprintf("Register %s with (%s)", modelConfig.Name, modelConfig.Schedule.String()))

		var scheduler *gocron.Scheduler
		if modelConfig.Schedule.Cron != "" {
			scheduler = mycron.Cron(modelConfig.Schedule.Cron)
		} else {
			scheduler = mycron.Every(modelConfig.Schedule.Every)
			if len(modelConfig.Schedule.At) > 0 {
				scheduler = scheduler.At(modelConfig.Schedule.At)
			} else {
				// If no $at present, delay start cron job with $eveny duration
				startDuration, _ := time.ParseDuration(modelConfig.Schedule.Every)
				scheduler = scheduler.StartAt(time.Now().Add(startDuration))
			}
		}

		scheduler.Do(func() {
			logger.Info("------------------------------------------------")
			logger.Info("performing", modelConfig.Name, "...")
			for _, modelConfig := range config.Models {
				m := model.Model{
					Config: modelConfig,
				}
				m.Perform()
			}
			logger.Info("------------------------------------------------")
		})
	}

	mycron.StartAsync()

	return nil
}

func Stop() {
	if mycron != nil {
		mycron.Stop()
	}
}
