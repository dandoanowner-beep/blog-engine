package email

import (
	"fmt"
	"net/smtp"
)

type SMTPSender struct {
	host string
	port string
	user string
	pass string
	from string
	appURL string
}

func NewSMTPSender(host, port, user, pass, from, appURL string) *SMTPSender {
	return &SMTPSender{host: host, port: port, user: user, pass: pass, from: from, appURL: appURL}
}

func (s *SMTPSender) SendVerification(to, token string) error {
	link := fmt.Sprintf("%s/auth/verify?token=%s", s.appURL, token)
	body := fmt.Sprintf("Subject: Verify your email\r\n\r\nClick to verify: %s", link)
	return s.send(to, body)
}

func (s *SMTPSender) SendPasswordReset(to, token string) error {
	link := fmt.Sprintf("%s/auth/reset-password?token=%s", s.appURL, token)
	body := fmt.Sprintf("Subject: Reset your password\r\n\r\nClick to reset: %s", link)
	return s.send(to, body)
}

func (s *SMTPSender) send(to, body string) error {
	auth := smtp.PlainAuth("", s.user, s.pass, s.host)
	return smtp.SendMail(
		s.host+":"+s.port,
		auth,
		s.from,
		[]string{to},
		[]byte(body),
	)
}
