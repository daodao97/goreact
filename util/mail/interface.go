package mail

var mailSender MailSender

func SetMailSender(sender MailSender) {
	mailSender = sender
}

func GetMailSender() MailSender {
	return mailSender
}

type MailSender interface {
	SendEmail(from, to string, subject string, plainTextContent string, htmlContent string) error
}
