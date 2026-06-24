package services

import (
	"fmt"
	"log"
	"mime"
	"net/smtp"
	"net/url"
	"strings"

	resend "github.com/resend/resend-go/v2"
)

type EmailSender interface {
	SendPasswordReset(email, token string) error
	SendEmailVerification(email, token string) error
	SendClientInvitation(email, clientName, token string) error
}

// ─── Resend ──────────────────────────────────────────────────────────────────

type resendEmailSender struct {
	client        *resend.Client
	from          string
	resetBaseURL  string
	verifyBaseURL string
	inviteBaseURL string
}

func (s *resendEmailSender) SendPasswordReset(email, token string) error {
	html, err := renderPasswordReset(buildURL(s.resetBaseURL, token))
	if err != nil {
		return fmt.Errorf("render password reset template: %w", err)
	}
	_, err = s.client.Emails.Send(&resend.SendEmailRequest{
		From:    s.from,
		To:      []string{email},
		Subject: "Réinitialisation de votre mot de passe",
		Html:    html,
	})
	return err
}

func (s *resendEmailSender) SendEmailVerification(email, token string) error {
	html, err := renderEmailVerification(buildURL(s.verifyBaseURL, token))
	if err != nil {
		return fmt.Errorf("render email verification template: %w", err)
	}
	_, err = s.client.Emails.Send(&resend.SendEmailRequest{
		From:    s.from,
		To:      []string{email},
		Subject: "Vérifiez votre adresse email",
		Html:    html,
	})
	return err
}

func (s *resendEmailSender) SendClientInvitation(email, clientName, token string) error {
	html, err := renderClientInvitation(buildURL(s.inviteBaseURL, token), clientName)
	if err != nil {
		return fmt.Errorf("render client invitation template: %w", err)
	}
	_, err = s.client.Emails.Send(&resend.SendEmailRequest{
		From:    s.from,
		To:      []string{email},
		Subject: "Vous avez été invité à rejoindre votre espace client",
		Html:    html,
	})
	return err
}

// ─── SMTP ────────────────────────────────────────────────────────────────────

type smtpEmailSender struct {
	host          string
	port          string
	user          string
	pass          string
	from          string
	resetBaseURL  string
	verifyBaseURL string
	inviteBaseURL string
}

func (s *smtpEmailSender) SendPasswordReset(email, token string) error {
	html, err := renderPasswordReset(buildURL(s.resetBaseURL, token))
	if err != nil {
		return fmt.Errorf("render password reset template: %w", err)
	}
	return s.sendMail(email, "Réinitialisation de votre mot de passe", html)
}

func (s *smtpEmailSender) SendEmailVerification(email, token string) error {
	html, err := renderEmailVerification(buildURL(s.verifyBaseURL, token))
	if err != nil {
		return fmt.Errorf("render email verification template: %w", err)
	}
	return s.sendMail(email, "Vérifiez votre adresse email", html)
}

func (s *smtpEmailSender) SendClientInvitation(email, clientName, token string) error {
	html, err := renderClientInvitation(buildURL(s.inviteBaseURL, token), clientName)
	if err != nil {
		return fmt.Errorf("render client invitation template: %w", err)
	}
	return s.sendMail(email, "Vous avez été invité à rejoindre votre espace client", html)
}

func (s *smtpEmailSender) sendMail(to, subject, htmlBody string) error {
	encodedSubject := mime.QEncoding.Encode("UTF-8", subject)
	msg := strings.Join([]string{
		fmt.Sprintf("From: %s", s.from),
		fmt.Sprintf("To: %s", to),
		fmt.Sprintf("Subject: %s", encodedSubject),
		"MIME-Version: 1.0",
		"Content-Type: text/html; charset=UTF-8",
		"",
		htmlBody,
	}, "\r\n")

	addr := fmt.Sprintf("%s:%s", s.host, s.port)
	var auth smtp.Auth
	if s.user != "" && s.pass != "" {
		auth = smtp.PlainAuth("", s.user, s.pass, s.host)
	}
	return smtp.SendMail(addr, auth, s.from, []string{to}, []byte(msg))
}

// ─── Log fallback ────────────────────────────────────────────────────────────

type logEmailSender struct {
	resetBaseURL  string
	verifyBaseURL string
	inviteBaseURL string
}

func (s *logEmailSender) SendPasswordReset(email, token string) error {
	log.Printf("password reset email fallback: to=%s reset_url=%s", email, buildURL(s.resetBaseURL, token))
	return nil
}

func (s *logEmailSender) SendEmailVerification(email, token string) error {
	log.Printf("email verification fallback: to=%s verify_url=%s", email, buildURL(s.verifyBaseURL, token))
	return nil
}

func (s *logEmailSender) SendClientInvitation(email, clientName, token string) error {
	log.Printf("client invitation email fallback: to=%s client=%s invite_url=%s", email, clientName, buildURL(s.inviteBaseURL, token))
	return nil
}

// ─── Factory ─────────────────────────────────────────────────────────────────

func NewEmailSenderFromEnv() EmailSender {
	resetBaseURL := getResetPasswordBaseURL()
	verifyBaseURL := getVerifyEmailBaseURL()
	inviteBaseURL := getClientInviteBaseURL()

	if apiKey := ResendAPIKey.GetValue(); apiKey != "" {
		from := SMTPFrom.GetValue()
		if from == "" {
			from = "no-reply@project-devis.local"
		}
		return &resendEmailSender{
			client:        resend.NewClient(apiKey),
			from:          from,
			resetBaseURL:  resetBaseURL,
			verifyBaseURL: verifyBaseURL,
			inviteBaseURL: inviteBaseURL,
		}
	}

	if host := SMTPHost.GetValue(); host != "" {
		port := SMTPPort.GetValue()
		if port == "" {
			port = "1025"
		}
		from := SMTPFrom.GetValue()
		if from == "" {
			from = "no-reply@project-devis.local"
		}
		return &smtpEmailSender{
			host:          host,
			port:          port,
			user:          SMTPUser.GetValue(),
			pass:          SMTPPassword.GetValue(),
			from:          from,
			resetBaseURL:  resetBaseURL,
			verifyBaseURL: verifyBaseURL,
			inviteBaseURL: inviteBaseURL,
		}
	}

	return &logEmailSender{resetBaseURL: resetBaseURL, verifyBaseURL: verifyBaseURL, inviteBaseURL: inviteBaseURL}
}

func getResetPasswordBaseURL() string {
	if v := ResetPasswordBaseURL.GetValue(); v != "" {
		return v
	}
	return "http://localhost:3000/reset-password"
}

func getVerifyEmailBaseURL() string {
	if v := VerifyEmailBaseURL.GetValue(); v != "" {
		return v
	}
	return "http://localhost:3000/verify-email"
}

func getClientInviteBaseURL() string {
	if v := ClientInviteBaseURL.GetValue(); v != "" {
		return v
	}
	return "http://localhost:3000/accept-invitation"
}

func buildURL(baseURL, token string) string {
	sep := "?"
	if strings.Contains(baseURL, "?") {
		sep = "&"
	}
	return baseURL + sep + "token=" + url.QueryEscape(token)
}
