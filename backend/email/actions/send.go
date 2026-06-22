package actions

import (
	"context"
	"log"

	emailGrpc "project-devis-email/services/grpc"
)

func (s *Server) insertEmailLog(ctx context.Context, userID, toEmail, emailType, referenceName, subject, resendID string, sendErr error) {
	status := "sent"
	if sendErr != nil {
		status = "failed"
	}
	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO email_logs (user_id, to_email, type, reference_name, subject, status, resend_id)
		 VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		nullableString(userID), toEmail, emailType, referenceName, subject, status, nullableString(resendID),
	); err != nil {
		log.Printf("insert email log failed: %v", err)
	}
}

func (s *Server) SendQuoteEmail(ctx context.Context, req *emailGrpc.SendQuoteEmailRequest) (*emailGrpc.SendEmailResponse, error) {
	if req.ToEmail == "" || req.QuoteName == "" {
		return &emailGrpc.SendEmailResponse{Success: false, Code: CodeInvalidInput}, nil
	}

	resendID, sendErr := s.sender.SendQuoteEmail(req.ToEmail, req.ToName, req.QuoteName, req.PdfBytes)
	if sendErr != nil {
		log.Printf("send quote email failed to=%s quote=%q: %v", req.ToEmail, req.QuoteName, sendErr)
	}

	subject := "Votre devis : " + req.QuoteName
	s.insertEmailLog(ctx, req.UserId, req.ToEmail, "quote_sent", req.QuoteName, subject, resendID, sendErr)

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
	if sendErr != nil {
		log.Printf("send schedule email failed to=%s quote=%q status=%s: %v", req.ToEmail, req.QuoteName, req.Status, sendErr)
	}

	emailType := "schedule_" + req.Status
	subject := scheduleSubjectFromStatus(req.QuoteName, req.Status)
	s.insertEmailLog(ctx, req.UserId, req.ToEmail, emailType, req.QuoteName, subject, resendID, sendErr)

	if sendErr != nil {
		return &emailGrpc.SendEmailResponse{Success: false, Code: CodeInternalError}, nil
	}
	return &emailGrpc.SendEmailResponse{Success: true, Code: CodeSuccess}, nil
}

func (s *Server) SendGenericEmail(ctx context.Context, req *emailGrpc.SendGenericEmailRequest) (*emailGrpc.SendEmailResponse, error) {
	if req.ToEmail == "" || req.Subject == "" {
		return &emailGrpc.SendEmailResponse{Success: false, Code: CodeInvalidInput}, nil
	}

	resendID, sendErr := s.sender.SendGenericEmail(
		req.ToEmail, req.ToName, req.Subject, req.TextBody,
		req.AttachmentName, req.AttachmentType, req.AttachmentBytes,
	)
	if sendErr != nil {
		log.Printf("send generic email failed to=%s subject=%q: %v", req.ToEmail, req.Subject, sendErr)
	}

	s.insertEmailLog(ctx, "", req.ToEmail, "generic", req.Subject, req.Subject, resendID, sendErr)

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
