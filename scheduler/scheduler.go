package scheduler

import (
	"fmt"
	"sync"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/gobackup/gobackup/config"
	superlogger "github.com/gobackup/gobackup/logger"
	"github.com/gobackup/gobackup/model"
)

var (
	mycron *gocron.Scheduler
)

// Start scheduler
func Start() error {
	logger := superlogger.Tag("Scheduler")

	mycron = gocron.NewScheduler(time.Local)

	mu := sync.Mutex{}

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

		scheduler.Do(func(modelConfig config.ModelConfig) {
			defer mu.Unlock()
			logger := superlogger.Tag(fmt.Sprintf("Scheduler: %s", modelConfig.Name))

			logger.Info("Performing...")

			m := model.Model{
				Config: modelConfig,
			}
			mu.Lock()
			m.Perform()
			logger.Info("Done.")
		}, modelConfig)
	}

	mycron.StartAsync()

	return nil
}

func Stop() {
	if mycron != nil {
		mycron.Stop()
	}
}
