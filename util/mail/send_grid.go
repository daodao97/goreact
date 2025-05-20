package mail

import (
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

type SendGridMailSender struct {
	apiKey string
	client *sendgrid.Client
}

func NewSendGridMailSender(apiKey string) MailSender {
	return &SendGridMailSender{
		apiKey: apiKey,
		client: sendgrid.NewSendClient(apiKey),
	}
}

func (s *SendGridMailSender) SendEmail(from, to string, subject string, plainTextContent string, htmlContent string) error {
	message := mail.NewSingleEmail(mail.NewEmail(from, from), subject, mail.NewEmail(to, to), plainTextContent, htmlContent)
	_, err := s.client.Send(message)
	return err
}
