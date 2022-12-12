package notifier

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gobackup/gobackup/logger"
)

type Webhook struct {
	Base

	Service string

	method          string
	contentType     string
	buildBody       func(title, message string) ([]byte, error)
	buildWebhookURL func(url string) (string, error)
	checkResult     func(status int, responseBody []byte) error
	buildHeaders    func() map[string]string
}

func (s *Webhook) getLogger() logger.Logger {
	return logger.Tag(fmt.Sprintf("Notifier: %s", s.Service))
}

func (s *Webhook) webhookURL() (string, error) {
	url := s.viper.GetString("url")

	if s.buildWebhookURL == nil {
		return url, nil
	}

	return s.buildWebhookURL(url)
}

func (s *Webhook) notify(title string, message string) error {
	logger := s.getLogger()

	s.viper.SetDefault("method", "POST")

	url, err := s.webhookURL()
	if err != nil {
		return err
	}

	payload, err := s.buildBody(title, message)
	if err != nil {
		return err
	}

	logger.Infof("Send message to %s...", url)
	req, err := http.NewRequest(s.method, url, strings.NewReader(string(payload)))
	if err != nil {
		logger.Error(err)
		return err
	}

	req.Header.Set("Content-Type", s.contentType)

	if s.buildHeaders != nil {
		headers := s.buildHeaders()
		for key, value := range headers {
			req.Header.Set(key, value)
		}
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		logger.Error(err)
		return err
	}
	defer resp.Body.Close()

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
