package mail

import (
	resend "github.com/resend/resend-go/v2"
)

// ResendMailSender implements MailSender using the Resend API.
type ResendMailSender struct {
    apiKey string
    client *resend.Client
}

// NewResendMailSender creates a new MailSender backed by Resend.
func NewResendMailSender(apiKey string) MailSender {
    return &ResendMailSender{
        apiKey: apiKey,
        client: resend.NewClient(apiKey),
    }
}

// SendEmail sends an email via Resend.
func (s *ResendMailSender) SendEmail(from, to string, subject string, plainTextContent string, htmlContent string) error {
    params := &resend.SendEmailRequest{
        From:    from,
        To:      []string{to},
        Subject: subject,
    }

    if htmlContent != "" {
        params.Html = htmlContent
    }
    if plainTextContent != "" {
        params.Text = plainTextContent
    }

    _, err := s.client.Emails.Send(params)
    return err
}
