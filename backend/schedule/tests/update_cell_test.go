package tests

import (
	"context"
	"testing"

	"project-devis-schedule/actions"
	scheduleGrpc "project-devis-schedule/services/grpc"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestUpdateScheduleCell_SuccessDraft(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT status FROM schedules WHERE schedule_id=\$1 AND user_id=\$2`).
		WithArgs("schedule-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(actions.StatusDraft))
	mock.ExpectExec(`UPDATE schedule_cells SET amount_cents=\$1, updated_at=NOW\(\) WHERE schedule_id=\$2 AND quote_line_id=\$3 AND month_index=\$4`).
		WithArgs(int64(1230), "schedule-1", "line-1", int32(2)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := srv.UpdateScheduleCell(context.Background(), &scheduleGrpc.UpdateScheduleCellRequest{
		ScheduleId:  "schedule-1",
		UserId:      "user-1",
		QuoteLineId: "line-1",
		MonthIndex:  2,
		AmountEur:   "12.30",
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

func TestUpdateScheduleCell_RejectedWhenFinalized(t *testing.T) {
	cases := []struct {
		name   string
		status string
	}{
		{name: "denied", status: actions.StatusDenied},
		{name: "valid", status: actions.StatusValid},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv, mock := setupServer(t)

			mock.ExpectQuery(`SELECT status FROM schedules WHERE schedule_id=\$1 AND user_id=\$2`).
				WithArgs("schedule-1", "user-1").
				WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(tc.status))

			resp, err := srv.UpdateScheduleCell(context.Background(), &scheduleGrpc.UpdateScheduleCellRequest{
				ScheduleId:  "schedule-1",
				UserId:      "user-1",
				QuoteLineId: "line-1",
				MonthIndex:  2,
				AmountEur:   "12.30",
			})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if resp.Success {
				t.Fatal("expected failure")
			}
			if resp.Code != actions.CodeScheduleFinalized {
				t.Fatalf("expected CodeScheduleFinalized, got %d", resp.Code)
			}
			if err := mock.ExpectationsWereMet(); err != nil {
				t.Fatalf("unmet expectations: %v", err)
			}
		})
	}
}

func TestUpdateScheduleCell_InvalidInput(t *testing.T) {
	cases := []struct {
		name      string
		monthIdx  int32
		amountEur string
	}{
		{name: "month index invalid", monthIdx: 0, amountEur: "12.30"},
		{name: "negative amount", monthIdx: 1, amountEur: "-1.00"},
		{name: "too many decimals", monthIdx: 1, amountEur: "12.345"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv, mock := setupServer(t)

			resp, err := srv.UpdateScheduleCell(context.Background(), &scheduleGrpc.UpdateScheduleCellRequest{
				ScheduleId:  "schedule-1",
				UserId:      "user-1",
				QuoteLineId: "line-1",
				MonthIndex:  tc.monthIdx,
				AmountEur:   tc.amountEur,
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
		})
	}
}

func TestUpdateScheduleCell_ScheduleNotFound(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT status FROM schedules WHERE schedule_id=\$1 AND user_id=\$2`).
		WithArgs("schedule-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"status"}))

	resp, err := srv.UpdateScheduleCell(context.Background(), &scheduleGrpc.UpdateScheduleCellRequest{
		ScheduleId:  "schedule-1",
		UserId:      "user-1",
		QuoteLineId: "line-1",
		MonthIndex:  2,
		AmountEur:   "12.30",
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

func TestUpdateScheduleCell_CellNotFound(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT status FROM schedules WHERE schedule_id=\$1 AND user_id=\$2`).
		WithArgs("schedule-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow(actions.StatusDraft))
	mock.ExpectExec(`UPDATE schedule_cells SET amount_cents=\$1, updated_at=NOW\(\) WHERE schedule_id=\$2 AND quote_line_id=\$3 AND month_index=\$4`).
		WithArgs(int64(1230), "schedule-1", "line-1", int32(2)).
		WillReturnResult(sqlmock.NewResult(0, 0))

	resp, err := srv.UpdateScheduleCell(context.Background(), &scheduleGrpc.UpdateScheduleCellRequest{
		ScheduleId:  "schedule-1",
		UserId:      "user-1",
		QuoteLineId: "line-1",
		MonthIndex:  2,
		AmountEur:   "12.30",
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
