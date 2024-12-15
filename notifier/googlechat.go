package notifier

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gobackup/gobackup/logger"
)

type GoogleChat struct {
	Base

	Service string

	method             string
	contentType        string
	buildBody          func(title, message string) ([]byte, error)
	buildGoogleChatURL func(url string) (string, error)
	checkResult        func(status int, responseBody []byte) error
	buildHeaders       func() map[string]string
}

type GoogleChatPayload struct {
	Text string `json:"text"`
}

func NewGoogleChat(base *Base) *GoogleChat {
	base.viper.SetDefault("method", "POST")

	return &GoogleChat{
		Base:        *base,
		Service:     "GoogleChat",
		method:      base.viper.GetString("method"),
		contentType: "application/json",
		buildBody: func(title, message string) ([]byte, error) {
			return json.Marshal(GoogleChatPayload{
				Text: title + "\r\n" + message,
			})
		},
		buildHeaders: func() map[string]string {
			headers := make(map[string]string)
			for key, value := range base.viper.GetStringMapString("headers") {
				headers[key] = value
			}

			return headers
		},
		checkResult: func(status int, responseBody []byte) error {
			if status == 200 {
				return nil
			}

			return fmt.Errorf("status: %d, body: %s", status, string(responseBody))
		},
	}
}

func (s *GoogleChat) getLogger() logger.Logger {
	return logger.Tag(fmt.Sprintf("Notifier: %s", s.Service))
}

func (s *GoogleChat) GoogleChatURL() (string, error) {
	url := s.viper.GetString("url")

	if s.buildGoogleChatURL == nil {
		return url, nil
	}

	return s.buildGoogleChatURL(url)
}

func (s *GoogleChat) notify(title string, message string) error {
	logger := s.getLogger()

	url, err := s.GoogleChatURL()
	if err != nil {
		return err
	}

	payload, err := s.buildBody(title, message)
	if err != nil {
		return err
	}

	logger.Infof("Send notification to %s...", url)
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
