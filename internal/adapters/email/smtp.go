package email

import (
	"fmt"
	"net/smtp"
)

// SMTPSender sends transactional emails via a plain SMTP relay.
// Auth is optional: if Username is empty the connection is unauthenticated
// (useful for local relay / MailHog).
type SMTPSender struct {
	host        string
	port        string
	username    string
	password    string
	fromAddress string
}

func NewSMTPSender(host, port, username, password, fromAddress string) *SMTPSender {
	return &SMTPSender{
		host:        host,
		port:        port,
		username:    username,
		password:    password,
		fromAddress: fromAddress,
	}
}

func (s *SMTPSender) SendVerificationEmail(toEmail, otp string) error {
	subject := "Verify your email address"
	body := fmt.Sprintf(
		"Hello,\r\n\r\n"+
			"Your email verification code is:\r\n\r\n"+
			"    %s\r\n\r\n"+
			"Enter this code to verify your email address. It expires in 24 hours.\r\n\r\n"+
			"If you did not create an account, you can safely ignore this email.\r\n",
		otp,
	)

	msg := buildMessage(s.fromAddress, toEmail, subject, body)
	addr := s.host + ":" + s.port

	var auth smtp.Auth
	if s.username != "" {
		auth = smtp.PlainAuth("", s.username, s.password, s.host)
	}

	return smtp.SendMail(addr, auth, s.fromAddress, []string{toEmail}, []byte(msg))
}

func buildMessage(from, to, subject, body string) string {
	return fmt.Sprintf(
		"From: %s\r\nTo: %s\r\nSubject: %s\r\nMIME-Version: 1.0\r\nContent-Type: text/plain; charset=\"utf-8\"\r\n\r\n%s",
		from, to, subject, body,
	)
}
