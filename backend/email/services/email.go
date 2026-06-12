package services

import (
	"fmt"
	"log"

	resend "github.com/resend/resend-go/v2"
)

type EmailSender interface {
	SendQuoteEmail(toEmail, toName, quoteName string, pdfBytes []byte) (resendID string, err error)
	SendScheduleEmail(toEmail, toName, quoteName, status string) (resendID string, err error)
}

// ─── Resend ──────────────────────────────────────────────────────────────────

type resendSender struct {
	client *resend.Client
	from   string
}

func (s *resendSender) SendQuoteEmail(toEmail, toName, quoteName string, pdfBytes []byte) (string, error) {
	html, err := RenderQuoteSent(toName, quoteName, len(pdfBytes) > 0)
	if err != nil {
		return "", fmt.Errorf("render quote sent template: %w", err)
	}
	req := &resend.SendEmailRequest{
		From:    s.from,
		To:      []string{toEmail},
		Subject: fmt.Sprintf("Votre devis : %s", quoteName),
		Html:    html,
	}
	if len(pdfBytes) > 0 {
		req.Attachments = []*resend.Attachment{
			{
				Filename: fmt.Sprintf("devis-%s.pdf", quoteName),
				Content:  pdfBytes,
			},
		}
	}
	resp, err := s.client.Emails.Send(req)
	if err != nil {
		return "", err
	}
	return resp.Id, nil
}

func (s *resendSender) SendScheduleEmail(toEmail, toName, quoteName, status string) (string, error) {
	html, err := RenderScheduleNotification(toName, quoteName, status)
	if err != nil {
		return "", fmt.Errorf("render schedule notification template: %w", err)
	}
	resp, err := s.client.Emails.Send(&resend.SendEmailRequest{
		From:    s.from,
		To:      []string{toEmail},
		Subject: scheduleSubject(quoteName, status),
		Html:    html,
	})
	if err != nil {
		return "", err
	}
	return resp.Id, nil
}

// ─── Log fallback ────────────────────────────────────────────────────────────

type logSender struct{}

func (s *logSender) SendQuoteEmail(toEmail, toName, quoteName string, _ []byte) (string, error) {
	log.Printf("email fallback: send quote to=%s quote=%q", toEmail, quoteName)
	return "", nil
}

func (s *logSender) SendScheduleEmail(toEmail, toName, quoteName, status string) (string, error) {
	log.Printf("email fallback: schedule notification to=%s quote=%q status=%s", toEmail, quoteName, status)
	return "", nil
}

// ─── Factory ─────────────────────────────────────────────────────────────────

func NewEmailSenderFromEnv() EmailSender {
	apiKey := ResendAPIKey.GetValue()
	if apiKey == "" {
		return &logSender{}
	}
	from := EmailFrom.GetValue()
	if from == "" {
		from = "no-reply@project-devis.local"
	}
	return &resendSender{
		client: resend.NewClient(apiKey),
		from:   from,
	}
}

// ─── Helpers ─────────────────────────────────────────────────────────────────

func scheduleSubject(quoteName, status string) string {
	switch status {
	case "VALID":
		return fmt.Sprintf("Échéancier validé — %s", quoteName)
	case "DENIED":
		return fmt.Sprintf("Échéancier refusé — %s", quoteName)
	default:
		return fmt.Sprintf("Mise à jour de l'échéancier — %s", quoteName)
	}
}
