package actions

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	emailGrpc "project-devis-email/services/grpc"
)

// mockEmailSender is a test double for services.EmailSender.
type mockEmailSender struct {
	sendQuoteFn    func(toEmail, toName, quoteName string, pdf []byte) (string, error)
	sendScheduleFn func(toEmail, toName, quoteName, status string) (string, error)
}

func (m *mockEmailSender) SendQuoteEmail(toEmail, toName, quoteName string, pdf []byte) (string, error) {
	if m.sendQuoteFn != nil {
		return m.sendQuoteFn(toEmail, toName, quoteName, pdf)
	}
	return "resend-id-123", nil
}

func (m *mockEmailSender) SendScheduleEmail(toEmail, toName, quoteName, status string) (string, error) {
	if m.sendScheduleFn != nil {
		return m.sendScheduleFn(toEmail, toName, quoteName, status)
	}
	return "resend-id-456", nil
}

func newTestServer(t *testing.T) (*Server, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return &Server{db: db, sender: &mockEmailSender{}}, mock
}

func TestSendQuoteEmail_Success(t *testing.T) {
	s, mock := newTestServer(t)

	// SQL has 'quote_sent' as a literal — 6 Go params: user_id, to_email, reference_name, subject, status, resend_id
	mock.ExpectExec(`INSERT INTO email_logs`).
		WithArgs(sqlmock.AnyArg(), "client@example.com", "quote_sent", "Devis Test", sqlmock.AnyArg(), "sent", "resend-id-123").
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := s.SendQuoteEmail(context.Background(), &emailGrpc.SendQuoteEmailRequest{
		ToEmail:   "client@example.com",
		ToName:    "Alice",
		QuoteName: "Devis Test",
		UserId:    "user-1",
		QuoteId:   "quote-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success=true, got code=%d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSendQuoteEmail_MissingEmail_ReturnsInvalidInput(t *testing.T) {
	s, _ := newTestServer(t)

	resp, err := s.SendQuoteEmail(context.Background(), &emailGrpc.SendQuoteEmailRequest{
		ToEmail:   "",
		QuoteName: "Devis Test",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success || resp.Code != CodeInvalidInput {
		t.Fatalf("expected invalid input, got success=%v code=%d", resp.Success, resp.Code)
	}
}

func TestSendQuoteEmail_SendFails_LogsFailedStatus(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	s := &Server{
		db: db,
		sender: &mockEmailSender{
			sendQuoteFn: func(_, _, _ string, _ []byte) (string, error) {
				return "", errors.New("resend API error")
			},
		},
	}

	mock.ExpectExec(`INSERT INTO email_logs`).
		WithArgs(sqlmock.AnyArg(), "client@example.com", "quote_sent", "Devis Test", sqlmock.AnyArg(), "failed", nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := s.SendQuoteEmail(context.Background(), &emailGrpc.SendQuoteEmailRequest{
		ToEmail:   "client@example.com",
		QuoteName: "Devis Test",
		UserId:    "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected gRPC error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected success=false when send fails")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestSendScheduleEmail_Valid(t *testing.T) {
	s, mock := newTestServer(t)

	// SQL type 'schedule_VALID' is a Go variable (not a literal) — 7 params
	mock.ExpectExec(`INSERT INTO email_logs`).
		WithArgs(sqlmock.AnyArg(), "client@example.com", "schedule_VALID", "Devis Test", sqlmock.AnyArg(), "sent", "resend-id-456").
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := s.SendScheduleEmail(context.Background(), &emailGrpc.SendScheduleEmailRequest{
		ToEmail:   "client@example.com",
		ToName:    "Bob",
		QuoteName: "Devis Test",
		Status:    "VALID",
		UserId:    "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success=true, got code=%d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestTrackEmailEvent_UpdatesOpened(t *testing.T) {
	s, mock := newTestServer(t)

	mock.ExpectExec(`UPDATE email_logs`).
		WithArgs("resend-abc", "email.opened", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := s.TrackEmailEvent(context.Background(), &emailGrpc.TrackEmailEventRequest{
		ResendId:   "resend-abc",
		EventType:  "email.opened",
		OccurredAt: "2026-06-09T10:00:00Z",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success=true, got code=%d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestTrackEmailEvent_MissingResendID(t *testing.T) {
	s, _ := newTestServer(t)

	resp, err := s.TrackEmailEvent(context.Background(), &emailGrpc.TrackEmailEventRequest{
		ResendId:  "",
		EventType: "email.opened",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success || resp.Code != CodeInvalidInput {
		t.Fatalf("expected invalid input, got success=%v code=%d", resp.Success, resp.Code)
	}
}
