package notifier

import (
	"fmt"
	"strings"

	"gopkg.in/gomail.v2"
)

type Mail struct {
	// Base is the base notifier
	from     string
	to       []string
	username string
	password string
	host     string
	port     int
}

func NewMail(base *Base) (*Mail, error) {
	base.viper.SetDefault("port", 25)

	username := base.viper.GetString("username")
	if username == "" {
		return nil, fmt.Errorf("username is required for mail notifier")
	}

	from := base.viper.GetString("from")
	if from == "" {
		from = username
	}

	port := base.viper.GetInt("port")

	return &Mail{
		username: username,
		password: base.viper.GetString("password"),
		to:       strings.Split(base.viper.GetString("to"), ","),
		from:     from,
		host:     base.viper.GetString("host"),
		port:     port,
	}, nil
}

func (s *Mail) notify(subject string, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", s.from)
	m.SetHeader("To", s.to...)
	m.SetHeader("Subject", subject)
	m.SetBody(`text/plain; charset="utf-8"`, body)

	d := gomail.NewDialer(s.host, s.port, s.username, s.password)

	if err := d.DialAndSend(m); err != nil {
		return fmt.Errorf("failed to send email: %v", err)
	}

	return nil
}
