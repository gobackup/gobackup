package notifier

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ses"
	"github.com/gobackup/gobackup/logger"
)

type SES struct {
	Base
	client *ses.SES
}

// type: ses
// region: us-east-1
// access_key_id: your-access-key-id
// secret_access_key: your-secret-access-key
func NewSES(base *Base) *SES {
	cfg := aws.NewConfig()
	base.viper.SetDefault("region", "us-east-1")

	cfg.Credentials = credentials.NewStaticCredentials(
		base.viper.GetString("access_key_id"),
		base.viper.GetString("secret_access_key"),
		base.viper.GetString("token"),
	)
	cfg.Region = aws.String(base.viper.GetString("region"))

	sess := session.Must(session.NewSession(cfg))
	return &SES{
		Base:   *base,
		client: ses.New(sess, cfg),
	}
}

func (s *SES) buildEmail(title, message string) *ses.SendEmailInput {
	return &ses.SendEmailInput{
		Destination: &ses.Destination{
			ToAddresses: []*string{
				aws.String(s.viper.GetString("to")),
			},
		},
		Message: &ses.Message{
			Body: &ses.Body{
				Text: &ses.Content{
					Charset: aws.String("UTF-8"),
					Data:    aws.String(message),
				},
			},
			Subject: &ses.Content{
				Charset: aws.String("UTF-8"),
				Data:    aws.String(title),
			},
		},
		Source: aws.String(s.viper.GetString("from")),
	}
}

func (s *SES) notify(title string, message string) error {
	logger := logger.Tag("Notifier: SES")

	logger.Info("Sending notification...")
	_, err := s.client.SendEmail(s.buildEmail(title, message))
	if err != nil {
		return err
	}
	logger.Info("Notification sent.")

	return nil
}
