package notifier

import (
	"encoding/json"
	"fmt"
	"github.com/gobackup/gobackup/logger"
	"io"
	"net/http"
	"strings"
)

type Healthchecks struct {
	Base
	Service         string
	successEndpoint string
	failureEndpoint string
	checkResult     func(status int, responseBody []byte) error
}

// NewHealthchecks
// type: healthchecks
// url: https://healthchecks.io/ping/e9ef3495-aa2b-4e8b-8927-a7138dd6c04c
func NewHealthchecks(base *Base) *Healthchecks {
	url := base.viper.GetString("url")

	return &Healthchecks{
		Base:            *base,
		Service:         "Healthchecks",
		successEndpoint: url,
		failureEndpoint: url + "/fail",
		checkResult: func(status int, responseBody []byte) error {
			if status == 200 {
				return nil
			}

			return fmt.Errorf("status: %d, body: %s", status, string(responseBody))
		},
	}
}

func (s *Healthchecks) getLogger() logger.Logger {
	return logger.Tag(fmt.Sprintf("Notifier: %s", s.Service))
}

func (s *Healthchecks) notify(title string, message string) error {
	logger := s.getLogger()
	url := s.failureEndpoint
	if strings.Contains(title, "OK") {
		url = s.successEndpoint
	}

	payload, err := json.Marshal(map[string]string{
		"title":   title,
		"message": message,
	})

	logger.Infof("Send notification to %s...", url)
	req, err := http.NewRequest(http.MethodPost, url, strings.NewReader(string(payload)))
	if err != nil {
		logger.Error(err)
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(err)
		return err
	}
	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			logger.Error(err)
		}
	}(resp.Body)

	var body []byte
	if resp.Body != nil {
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return err
		}
	}

	if s.checkResult != nil {
		err = s.checkResult(resp.StatusCode, body)
		if err != nil {
			logger.Error(err)
			return nil
		}
	} else {
		logger.Infof("Response body: %s", string(body))
	}

	logger.Info("Notification sent.")

	return nil
}
