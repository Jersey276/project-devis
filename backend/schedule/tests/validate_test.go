package tests

import (
	"context"
	"testing"

	"project-devis-schedule/actions"
	scheduleGrpc "project-devis-schedule/services/grpc"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestValidateSchedule_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT quote_id, status FROM schedules WHERE schedule_id=\$1 AND user_id=\$2 FOR UPDATE`).
		WithArgs("schedule-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"quote_id", "status"}).AddRow("quote-1", actions.StatusDraft))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM schedules WHERE quote_id=\$1 AND schedule_id<>\$2 AND status='VALID'`).
		WithArgs("quote-1", "schedule-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM \(\s*SELECT sc\.quote_line_id`).
		WithArgs("schedule-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	mock.ExpectExec(`UPDATE schedules SET status='VALID', validated_at=NOW\(\), updated_at=NOW\(\) WHERE schedule_id=\$1 AND user_id=\$2`).
		WithArgs("schedule-1", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectExec(`UPDATE schedules SET status='DENIED', updated_at=NOW\(\) WHERE quote_id=\$1 AND schedule_id<>\$2 AND status IN \('DRAFT','NEGOCIATE'\)`).
		WithArgs("quote-1", "schedule-1").
		WillReturnResult(sqlmock.NewResult(0, 2))
	mock.ExpectCommit()

	resp, err := srv.ValidateSchedule(context.Background(), &scheduleGrpc.ValidateScheduleRequest{
		ScheduleId: "schedule-1",
		UserId:     "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestValidateSchedule_InvalidInput(t *testing.T) {
	srv, mock := setupServer(t)

	resp, err := srv.ValidateSchedule(context.Background(), &scheduleGrpc.ValidateScheduleRequest{
		ScheduleId: "",
		UserId:     "user-1",
	})
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

func TestValidateSchedule_NotFound(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT quote_id, status FROM schedules WHERE schedule_id=\$1 AND user_id=\$2 FOR UPDATE`).
		WithArgs("schedule-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"quote_id", "status"}))
	mock.ExpectRollback()

	resp, err := srv.ValidateSchedule(context.Background(), &scheduleGrpc.ValidateScheduleRequest{
		ScheduleId: "schedule-1",
		UserId:     "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure")
	}
	if resp.Code != actions.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestValidateSchedule_AlreadyValidated(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT quote_id, status FROM schedules WHERE schedule_id=\$1 AND user_id=\$2 FOR UPDATE`).
		WithArgs("schedule-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"quote_id", "status"}).AddRow("quote-1", actions.StatusValid))
	mock.ExpectRollback()

	resp, err := srv.ValidateSchedule(context.Background(), &scheduleGrpc.ValidateScheduleRequest{
		ScheduleId: "schedule-1",
		UserId:     "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure")
	}
	if resp.Code != actions.CodeScheduleValidated {
		t.Fatalf("expected CodeScheduleValidated, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestValidateSchedule_Unbalanced(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT quote_id, status FROM schedules WHERE schedule_id=\$1 AND user_id=\$2 FOR UPDATE`).
		WithArgs("schedule-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"quote_id", "status"}).AddRow("quote-1", actions.StatusNegotiate))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM schedules WHERE quote_id=\$1 AND schedule_id<>\$2 AND status='VALID'`).
		WithArgs("quote-1", "schedule-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM \(\s*SELECT sc\.quote_line_id`).
		WithArgs("schedule-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectRollback()

	resp, err := srv.ValidateSchedule(context.Background(), &scheduleGrpc.ValidateScheduleRequest{
		ScheduleId: "schedule-1",
		UserId:     "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure")
	}
	if resp.Code != actions.CodeScheduleUnbalanced {
		t.Fatalf("expected CodeScheduleUnbalanced, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestValidateSchedule_AnotherValidExists(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT quote_id, status FROM schedules WHERE schedule_id=\$1 AND user_id=\$2 FOR UPDATE`).
		WithArgs("schedule-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"quote_id", "status"}).AddRow("quote-1", actions.StatusDraft))
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM schedules WHERE quote_id=\$1 AND schedule_id<>\$2 AND status='VALID'`).
		WithArgs("quote-1", "schedule-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))
	mock.ExpectRollback()

	resp, err := srv.ValidateSchedule(context.Background(), &scheduleGrpc.ValidateScheduleRequest{
		ScheduleId: "schedule-1",
		UserId:     "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure")
	}
	if resp.Code != actions.CodeScheduleValidated {
		t.Fatalf("expected CodeScheduleValidated, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}