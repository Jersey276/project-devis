package services

import (
	"fmt"
	"log"
	"net/smtp"
	"net/url"
	"strings"
)

type EmailSender interface {
	SendPasswordReset(email, token string) error
}

type smtpEmailSender struct {
	host    string
	port    string
	user    string
	pass    string
	from    string
	baseURL string
}

type logEmailSender struct {
	baseURL string
}

func NewEmailSenderFromEnv() EmailSender {
	host := SMTPHost.GetValue()
	if host == "" {
		return &logEmailSender{baseURL: getResetPasswordBaseURL()}
	}

	port := SMTPPort.GetValue()
	if port == "" {
		port = "1025"
	}

	from := SMTPFrom.GetValue()
	if from == "" {
		from = "no-reply@project-devis.local"
	}

	return &smtpEmailSender{
		host:    host,
		port:    port,
		user:    SMTPUser.GetValue(),
		pass:    SMTPPassword.GetValue(),
		from:    from,
		baseURL: getResetPasswordBaseURL(),
	}
}

func (s *smtpEmailSender) SendPasswordReset(email, token string) error {
	resetURL := buildResetURL(s.baseURL, token)
	subject := "Réinitialisation de votre mot de passe"
	body := fmt.Sprintf("Bonjour,\r\n\r\nPour réinitialiser votre mot de passe, utilisez ce lien:\r\n%s\r\n\r\nCe lien expire dans 15 minutes.\r\n", resetURL)

	msg := strings.Join([]string{
		fmt.Sprintf("From: %s", s.from),
		fmt.Sprintf("To: %s", email),
		fmt.Sprintf("Subject: %s", subject),
		"MIME-Version: 1.0",
		"Content-Type: text/plain; charset=UTF-8",
		"",
		body,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	var auth smtp.Auth
	if s.user != "" && s.pass != "" {
		auth = smtp.PlainAuth("", s.user, s.pass, s.host)
	}

	return smtp.SendMail(addr, auth, s.from, []string{email}, []byte(msg))
}

func (s *logEmailSender) SendPasswordReset(email, token string) error {
	resetURL := buildResetURL(s.baseURL, token)
	log.Printf("password reset email fallback: to=%s reset_url=%s", email, resetURL)
	return nil
}

func getResetPasswordBaseURL() string {
	baseURL := ResetPasswordBaseURL.GetValue()
	if baseURL == "" {
		baseURL = "http://localhost:3000/reset-password"
	}
	return baseURL
}

func buildResetURL(baseURL, token string) string {
	separator := "?"
	if strings.Contains(baseURL, "?") {
		separator = "&"
	}
	return baseURL + separator + "token=" + url.QueryEscape(token)
}
