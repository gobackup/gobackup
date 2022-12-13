package notifier

import (
	"encoding/base64"
	"fmt"
	"net/smtp"
	"sort"
	"strings"

	"github.com/gobackup/gobackup/logger"
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
	logger := logger.Tag("Notifier: mail")
	mail := &Mail{}

	base.viper.SetDefault("port", "25")

	mail.username = base.viper.GetString("username")
	mail.password = base.viper.GetString("password")

	mail.to = strings.Split(base.viper.GetString("to"), ",")
	if len(base.viper.GetString("username")) == 0 {
		logger.Fatalf("username is required for mail notifier")
	}

	mail.from = base.viper.GetString("from")
	if len(mail.from) == 0 {
		mail.from = mail.username
	}

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
	headers := make(map[string]string)
	headers["From"] = s.from
	headers["To"] = strings.Join(s.to, ",")
	headers["Subject"] = title
	headers["Content-Type"] = `text/plain; charset="utf-8"`
	headers["Content-Transfer-Encoding"] = "base64"

	// Sort headers by key
	var keys []string
	for k := range headers {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	headerTexts := []string{}

	for _, k := range keys {
		headerTexts = append(headerTexts, fmt.Sprintf("%s: %s", k, headers[k]))
	}

	return fmt.Sprintf("%s\n%s", strings.Join(headerTexts, "\n"), base64.StdEncoding.EncodeToString([]byte(message)))
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
