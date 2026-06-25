package tests

import (
	"context"
	"fmt"
	"testing"
	"time"

	"project-devis-schedule/actions"
	scheduleGrpc "project-devis-schedule/services/grpc"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCreateSchedule_Success(t *testing.T) {
	srv, mock := setupServer(t)
	req := &scheduleGrpc.CreateScheduleRequest{
		UserId:         "user-1",
		QuoteId:        "quote-1",
		Name:           "Echeancier principal",
		StartMonth:     "2026-06",
		DurationMonths: 3,
	}
	lineIDs := []string{"line-1", "line-2"}

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO schedules`).
		WithArgs(sqlmock.AnyArg(), "quote-1", "user-1", "Echeancier principal", actions.StatusDraft, time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC), int32(3), nil).
		WillReturnResult(sqlmock.NewResult(1, 1))
	for _, lineID := range lineIDs {
		for monthIndex := 1; monthIndex <= int(req.DurationMonths); monthIndex++ {
			mock.ExpectExec(`INSERT INTO schedule_cells`).
				WithArgs(sqlmock.AnyArg(), lineID, monthIndex, int64(0)).
				WillReturnResult(sqlmock.NewResult(1, 1))
		}
	}
	mock.ExpectCommit()

	resp, err := srv.CreateScheduleWithEligibleLines(context.Background(), req, lineIDs)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.ScheduleId == "" {
		t.Fatal("expected non-empty schedule_id")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateSchedule_InvalidInput(t *testing.T) {
	srv, mock := setupServer(t)

	resp, err := srv.CreateScheduleWithEligibleLines(context.Background(), &scheduleGrpc.CreateScheduleRequest{
		UserId:         "user-1",
		QuoteId:        "quote-1",
		Name:           "",
		StartMonth:     "2026-06",
		DurationMonths: 3,
	}, []string{"line-1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for invalid input")
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB calls: %v", err)
	}
}

func TestCreateSchedule_NoEligibleLines(t *testing.T) {
	srv, mock := setupServer(t)

	resp, err := srv.CreateScheduleWithEligibleLines(context.Background(), &scheduleGrpc.CreateScheduleRequest{
		UserId:         "user-1",
		QuoteId:        "quote-1",
		Name:           "Echeancier principal",
		StartMonth:     "2026-06",
		DurationMonths: 3,
	}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure when no eligible lines are provided")
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB calls: %v", err)
	}
}

func TestCreateSchedule_RollsBackOnCellInsertError(t *testing.T) {
	srv, mock := setupServer(t)
	req := &scheduleGrpc.CreateScheduleRequest{
		UserId:         "user-1",
		QuoteId:        "quote-1",
		Name:           "Echeancier principal",
		StartMonth:     "2026-06",
		DurationMonths: 2,
	}

	mock.ExpectBegin()
	mock.ExpectExec(`INSERT INTO schedules`).
		WithArgs(sqlmock.AnyArg(), "quote-1", "user-1", "Echeancier principal", actions.StatusDraft, time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC), int32(2), nil).
		WillReturnResult(sqlmock.NewResult(1, 1))
	mock.ExpectExec(`INSERT INTO schedule_cells`).
		WithArgs(sqlmock.AnyArg(), "line-1", 1, int64(0)).
		WillReturnError(fmt.Errorf("insert failed"))
	mock.ExpectRollback()

	resp, err := srv.CreateScheduleWithEligibleLines(context.Background(), req, []string{"line-1"})
	if err == nil {
		t.Fatal("expected error")
	}
	if resp.Success {
		t.Fatal("expected failure when cell insert fails")
	}
	if resp.Code != actions.CodeInternalError {
		t.Fatalf("expected CodeInternalError, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
