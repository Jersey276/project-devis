package actions

import (
	"context"
	"log"

	emailGrpc "project-devis-email/services/grpc"
)

func (s *Server) SendQuoteEmail(ctx context.Context, req *emailGrpc.SendQuoteEmailRequest) (*emailGrpc.SendEmailResponse, error) {
	if req.ToEmail == "" || req.QuoteName == "" {
		return &emailGrpc.SendEmailResponse{Success: false, Code: CodeInvalidInput}, nil
	}

	resendID, sendErr := s.sender.SendQuoteEmail(req.ToEmail, req.ToName, req.QuoteName, req.PdfBytes)

	status := "sent"
	if sendErr != nil {
		log.Printf("send quote email failed to=%s quote=%q: %v", req.ToEmail, req.QuoteName, sendErr)
		status = "failed"
	}

	subject := "Votre devis : " + req.QuoteName
	_, dbErr := s.db.ExecContext(ctx,
		`INSERT INTO email_logs (user_id, to_email, type, reference_name, subject, status, resend_id)
		 VALUES ($1, $2, 'quote_sent', $3, $4, $5, $6)`,
		nullableString(req.UserId), req.ToEmail, req.QuoteName, subject, status, nullableString(resendID),
	)
	if dbErr != nil {
		log.Printf("insert email log failed: %v", dbErr)
	}

	if sendErr != nil {
		return &emailGrpc.SendEmailResponse{Success: false, Code: CodeInternalError}, nil
	}
	return &emailGrpc.SendEmailResponse{Success: true, Code: CodeSuccess}, nil
}

func (s *Server) SendScheduleEmail(ctx context.Context, req *emailGrpc.SendScheduleEmailRequest) (*emailGrpc.SendEmailResponse, error) {
	if req.ToEmail == "" || req.QuoteName == "" {
		return &emailGrpc.SendEmailResponse{Success: false, Code: CodeInvalidInput}, nil
	}

	resendID, sendErr := s.sender.SendScheduleEmail(req.ToEmail, req.ToName, req.QuoteName, req.Status)

	status := "sent"
	if sendErr != nil {
		log.Printf("send schedule email failed to=%s quote=%q status=%s: %v", req.ToEmail, req.QuoteName, req.Status, sendErr)
		status = "failed"
	}

	emailType := "schedule_" + req.Status
	subject := scheduleSubjectFromStatus(req.QuoteName, req.Status)
	_, dbErr := s.db.ExecContext(ctx,
		`INSERT INTO email_logs (user_id, to_email, type, reference_name, subject, status, resend_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		nullableString(req.UserId), req.ToEmail, emailType, req.QuoteName, subject, status, nullableString(resendID),
	)
	if dbErr != nil {
		log.Printf("insert email log failed: %v", dbErr)
	}

	if sendErr != nil {
		return &emailGrpc.SendEmailResponse{Success: false, Code: CodeInternalError}, nil
	}
	return &emailGrpc.SendEmailResponse{Success: true, Code: CodeSuccess}, nil
}

func nullableString(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}

func scheduleSubjectFromStatus(quoteName, status string) string {
	switch status {
	case "VALID":
		return "Échéancier validé — " + quoteName
	case "DENIED":
		return "Échéancier refusé — " + quoteName
	default:
		return "Mise à jour de l'échéancier — " + quoteName
	}
}
