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
	URL     string

	method      string
	contentType string
	buildBody   func(title, message string) ([]byte, error)
}

func (s *Webhook) getLogger() logger.Logger {
	return logger.Tag(fmt.Sprintf("Notifier: %s", s.Service))
}

func (s *Webhook) notify(title string, message string) error {
	logger := s.getLogger()

	s.viper.SetDefault("method", "POST")

	s.URL = s.viper.GetString("url")

	payload, err := s.buildBody(title, message)
	if err != nil {
		return err
	}

	logger.Infof("Send message to %s...", s.URL)
	req, err := http.NewRequest(s.method, s.URL, strings.NewReader(string(payload)))
	if err != nil {
		logger.Error(err)
		return err
	}

	req.Header.Set("Content-Type", s.contentType)

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

	logger.Debugf("Response body: %s", string(body))
	logger.Info("Notification sent.")

	return nil
}
