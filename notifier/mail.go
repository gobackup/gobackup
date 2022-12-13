package notifier

import (
	"fmt"
	"net/smtp"
	"strings"
)

type Mail struct {
	// Base is the base notifier
	from     string
	to       []string
	username string
	password string
	host     string
	port     string
}

func NewMail(base *Base) *Mail {
	mail := &Mail{}

	base.viper.SetDefault("port", "25")

	mail.from = base.viper.GetString("from")
	mail.to = strings.Split(base.viper.GetString("to"), ",")
	if len(base.viper.GetString("username")) > 0 {
		mail.username = base.viper.GetString("username")
	} else {
		mail.username = mail.from
	}
	mail.password = base.viper.GetString("password")
	mail.host = base.viper.GetString("host")
	mail.port = base.viper.GetString("port")

	return mail
}

func (s Mail) getAddr() string {
	return fmt.Sprintf("%s:%s", s.host, s.port)
}

func (s Mail) getAuth() smtp.Auth {
	return smtp.PlainAuth("", s.username, s.password, s.host)
}

func (s Mail) buildBody(title string, message string) string {
	return fmt.Sprintf("To: %s\r\n"+
		"Subject: %s\r\n"+
		"\r\n"+
		"%s\r\n", strings.Join(s.to, ","), title, message)
}

func (s *Mail) notify(title string, message string) error {
	auth := s.getAuth()

	// Connect to the server, authenticate, set the sender and recipient,
	// and send the email all in one step.
	to := s.to

	err := smtp.SendMail(s.getAddr(), auth, s.from, to, []byte(s.buildBody(title, message)))
	if err != nil {
		return err
	}

	return nil
}
