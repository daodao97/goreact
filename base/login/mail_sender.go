package login

import (
	"github.com/sendgrid/sendgrid-go"
	"github.com/sendgrid/sendgrid-go/helpers/mail"
)

var mailSender MailSender

func SetMailSender(sender MailSender) {
	mailSender = sender
}

func GetMailSender() MailSender {
	return mailSender
}

type MailSender interface {
	SendVerificationCode(to string, subject string, plainTextContent string) error
}

type SendGridMailSender struct {
	apiKey string
	from   string
}

func NewSendGridMailSender(apiKey string, from string) MailSender {
	return &SendGridMailSender{
		apiKey: apiKey,
		from:   from,
	}
}

func (s *SendGridMailSender) SendVerificationCode(to string, subject string, plainTextContent string) error {
	client := sendgrid.NewSendClient(s.apiKey)
	message := mail.NewSingleEmail(mail.NewEmail(s.from, s.from), subject, mail.NewEmail(to, to), plainTextContent, "")
	_, err := client.Send(message)
	return err
}
