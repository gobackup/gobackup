package notifier

import (
	"fmt"
	"time"

	"github.com/gobackup/gobackup/config"
	"github.com/gobackup/gobackup/logger"
	"github.com/spf13/viper"
)

type Base struct {
	viper     *viper.Viper
	Name      string
	onSuccess bool
	onFailure bool
}

type Notifier interface {
	notify(title, message string) error
}

var (
	notifyTypeSuccess = 1
	notifyTypeFailure = 2
)

func newNotifier(name string, config config.SubConfig) (Notifier, *Base, error) {
	base := &Base{
		viper: config.Viper,
		Name:  name,
	}
	base.viper.SetDefault("on_success", true)
	base.viper.SetDefault("on_failure", true)

	base.onSuccess = base.viper.GetBool("on_success")
	base.onFailure = base.viper.GetBool("on_failure")

	switch config.Type {
	case "webhook":
		return &Webhook{Base: *base}, base, nil
	case "feishu":
		return NewFeishu(base), base, nil
	case "dingtalk":
		return NewDingtalk(base), base, nil
	case "discord":
		return NewDiscord(base), base, nil
	}

	return nil, nil, fmt.Errorf("Notifier: %s is not supported", name)
}

func notify(model config.ModelConfig, title, message string, notifyType int) error {
	logger := logger.Tag("Notifier")

	logger.Infof("Running %d Notifiers", len(model.Notifiers))
	for name, config := range model.Notifiers {
		notifier, base, err := newNotifier(name, config)
		if err != nil {
			logger.Error(err)
			continue
		}

		if notifyType == notifyTypeSuccess {
			if base.onSuccess {
				if err := notifier.notify(title, message); err != nil {
					logger.Error(err)
				}
			}
		} else if notifyType == notifyTypeFailure {
			if base.onFailure {
				if err := notifier.notify(title, message); err != nil {
					logger.Error(err)
				}
			}
		}
	}

	return nil
}

func Success(model config.ModelConfig) error {
	title := fmt.Sprintf("[GoBackup] Backup of %s completed successfully", model.Name)
	message := fmt.Sprintf("Backup of %s completed successfully at %s", model.Name, time.Now().Local())
	return notify(model, title, message, notifyTypeSuccess)
}

func Failure(model config.ModelConfig, reason string) error {
	title := fmt.Sprintf("[GoBackup] Backup of %s failed", model.Name)
	message := fmt.Sprintf("Backup of %s failed at %s:\n\n%s", model.Name, time.Now().Local(), reason)

	return notify(model, title, message, notifyTypeFailure)
}
