package tests

import (
	"context"
	"testing"
	"time"

	"project-devis-schedule/actions"
	scheduleGrpc "project-devis-schedule/services/grpc"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestListSchedules_ByQuote_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM schedules WHERE user_id = \$1 AND quote_id = \$2`).
		WithArgs("user-1", "quote-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(2))

	mock.ExpectQuery(`SELECT schedule_id, quote_id, status, name, start_month, duration_months\s+FROM schedules WHERE user_id = \$1 AND quote_id = \$2 ORDER BY created_at DESC LIMIT \$3 OFFSET \$4`).
		WithArgs("user-1", "quote-1", int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"schedule_id", "quote_id", "status", "name", "start_month", "duration_months"}).
			AddRow("schedule-2", "quote-1", actions.StatusNegotiate, "Plan B", time.Date(2026, time.July, 1, 0, 0, 0, 0, time.UTC), int32(4)).
			AddRow("schedule-1", "quote-1", actions.StatusDraft, "Plan A", time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC), int32(3)))

	resp, err := srv.ListSchedules(context.Background(), &scheduleGrpc.ListSchedulesRequest{
		UserId:  "user-1",
		QuoteId: "quote-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if len(resp.Schedules) != 2 {
		t.Fatalf("expected 2 schedules, got %d", len(resp.Schedules))
	}
	if resp.Total != 2 {
		t.Fatalf("expected total 2, got %d", resp.Total)
	}
	if resp.Schedules[0].ScheduleId != "schedule-2" {
		t.Fatalf("expected first schedule schedule-2, got %q", resp.Schedules[0].ScheduleId)
	}
	if resp.Schedules[0].StartMonth != "2026-07" {
		t.Fatalf("expected start_month 2026-07, got %q", resp.Schedules[0].StartMonth)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestListSchedules_ByUser_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM schedules WHERE user_id = \$1`).
		WithArgs("user-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectQuery(`SELECT schedule_id, quote_id, status, name, start_month, duration_months\s+FROM schedules WHERE user_id = \$1 ORDER BY created_at DESC LIMIT \$2 OFFSET \$3`).
		WithArgs("user-1", int32(20), int32(0)).
		WillReturnRows(sqlmock.NewRows([]string{"schedule_id", "quote_id", "status", "name", "start_month", "duration_months"}).
			AddRow("schedule-9", "quote-9", actions.StatusValid, "Plan valide", time.Date(2026, time.September, 1, 0, 0, 0, 0, time.UTC), int32(2)))

	resp, err := srv.ListSchedules(context.Background(), &scheduleGrpc.ListSchedulesRequest{UserId: "user-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if len(resp.Schedules) != 1 {
		t.Fatalf("expected 1 schedule, got %d", len(resp.Schedules))
	}
	if resp.Total != 1 {
		t.Fatalf("expected total 1, got %d", resp.Total)
	}
	if resp.Schedules[0].QuoteId != "quote-9" {
		t.Fatalf("expected quote-9, got %q", resp.Schedules[0].QuoteId)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestListSchedules_InvalidInput(t *testing.T) {
	srv, mock := setupServer(t)

	resp, err := srv.ListSchedules(context.Background(), &scheduleGrpc.ListSchedulesRequest{UserId: ""})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure")
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB calls: %v", err)
	}
}
