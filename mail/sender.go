package mail

import (
	"fmt"
	"github.com/jordan-wright/email"
	"net/smtp"
)

// only for production
const (
	smtpAuthAddress   = "smtp.gmail.com"
	smtpServerAddress = "smtp.gmail.com:587"
)

type EmailSender interface {
	SendEmail(
		subject string,
		content string,
		to []string,
		cc []string,
		bcc []string,
		attachFile []string,
	) error
}

type GmailSender struct {
	name              string
	fromEmailAddress  string
	fromEmailPassword string
}

func NewGmailSender(name string, fromEmailAddress string, fromEmailPassword string) EmailSender {
	return &GmailSender{
		name:              name,
		fromEmailAddress:  fromEmailAddress,
		fromEmailPassword: fromEmailPassword,
	}
}

func (g *GmailSender) SendEmail(subject string, content string, to []string, cc []string, bcc []string, attachFile []string) error {
	fmt.Println("SendEmailfunction called.")
	e := email.NewEmail()
	e.From = fmt.Sprintf("%s <%s>", g.name, g.fromEmailAddress)
	e.Subject = subject
	e.HTML = []byte(content)
	e.Bcc = bcc
	e.Cc = cc
	e.To = to

	for _, f := range attachFile {
		_, err := e.AttachFile(f)
		if err != nil {
			return fmt.Errorf("fail to attach file: %s %w\n", f, err)
		}
	}

	//smtpAuth := smtp.PlainAuth("", g.fromEmailAddress, g.fromEmailPassword, smtpAuthAddress)
	//return e.Send(smtpServerAddress, smtpAuth)

	smtpAuth := smtp.PlainAuth("", "", "", "localhost")
	return e.Send("localhost:1025", smtpAuth)
}
