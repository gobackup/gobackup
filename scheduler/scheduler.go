package scheduler

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/go-co-op/gocron"
	"github.com/gobackup/gobackup/config"
	superlogger "github.com/gobackup/gobackup/logger"
	"github.com/gobackup/gobackup/model"
)

var (
	mycron *gocron.Scheduler

	// Regex to match duration strings with extended units like "1day", "2weeks", etc.
	extendedDurationRegex = regexp.MustCompile(`^(\d+)\s*(day|days|d|week|weeks|w|month|months)$`)
)

// parseDuration parses a duration string, supporting extended units like "day", "week", "month"
// in addition to Go's standard time.ParseDuration units.
func parseDuration(s string) (time.Duration, error) {
	// First try Go's standard ParseDuration
	if d, err := time.ParseDuration(s); err == nil {
		return d, nil
	}

	// Try to match extended units (case-insensitive)
	matches := extendedDurationRegex.FindStringSubmatch(strings.ToLower(s))
	if matches == nil {
		return 0, fmt.Errorf("invalid duration format: %s", s)
	}

	value, _ := strconv.Atoi(matches[1])
	unit := matches[2]

	switch unit {
	case "day", "days", "d":
		return time.Duration(value) * 24 * time.Hour, nil
	case "week", "weeks", "w":
		return time.Duration(value) * 7 * 24 * time.Hour, nil
	case "month", "months":
		// Approximate month as 30 days
		return time.Duration(value) * 30 * 24 * time.Hour, nil
	}

	return 0, fmt.Errorf("invalid duration format: %s", s)
}

func init() {
	config.OnConfigChange(func(in fsnotify.Event) {
		Restart()
	})
}

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
				// If no $at present, delay start cron job with $every duration
				startDuration, _ := parseDuration(modelConfig.Schedule.Every)
				scheduler = scheduler.StartAt(time.Now().Add(startDuration))
			}
		}

		if _, err := scheduler.Do(func(modelConfig config.ModelConfig) {
			defer mu.Unlock()
			logger := superlogger.Tag(fmt.Sprintf("Scheduler: %s", modelConfig.Name))

			logger.Info("Performing...")

			m := model.Model{
				Config: modelConfig,
			}
			mu.Lock()
			if err := m.Perform(); err != nil {
				logger.Errorf("Failed to perform: %s", err.Error())
			}
			logger.Info("Done.")
		}, modelConfig); err != nil {
			logger.Errorf("Failed to register job func: %s", err.Error())
		}
	}

	mycron.StartAsync()

	return nil
}

func Restart() error {
	logger := superlogger.Tag("Scheduler")
	logger.Info("Reloading...")
	Stop()
	return Start()
}

func Stop() {
	if mycron != nil {
		mycron.Stop()
	}
}
