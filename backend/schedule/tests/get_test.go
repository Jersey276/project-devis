package tests

import (
	"context"
	"testing"
	"time"

	"project-devis-schedule/actions"
	scheduleGrpc "project-devis-schedule/services/grpc"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetSchedule_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT quote_id, status, name, start_month, duration_months FROM schedules WHERE schedule_id=\$1 AND user_id=\$2`).
		WithArgs("schedule-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"quote_id", "status", "name", "start_month", "duration_months"}).
			AddRow("quote-1", actions.StatusDraft, "Echeancier principal", time.Date(2026, time.June, 1, 0, 0, 0, 0, time.UTC), int32(2)))

	mock.ExpectQuery(`SELECT sc.quote_line_id, COALESCE\(SUM\(sc.amount_cents\), 0\), COALESCE\(ROUND\(ql.unit_price \* ql.quantity\), 0\)::BIGINT`).
		WithArgs("schedule-1").
		WillReturnRows(sqlmock.NewRows([]string{"quote_line_id", "planned_cents", "line_cents"}).
			AddRow("line-1", int64(900), int64(1000)).
			AddRow("line-2", int64(500), int64(500)))

	mock.ExpectQuery(`SELECT month_index, COALESCE\(SUM\(amount_cents\), 0\) FROM schedule_cells WHERE schedule_id=\$1 GROUP BY month_index ORDER BY month_index`).
		WithArgs("schedule-1").
		WillReturnRows(sqlmock.NewRows([]string{"month_index", "amount_cents"}).
			AddRow(int32(1), int64(700)).
			AddRow(int32(2), int64(700)))

	mock.ExpectQuery(`SELECT COALESCE\(SUM\(ROUND\(unit_price \* quantity\)\), 0\)::BIGINT FROM quote_lines WHERE quote_id=\$1`).
		WithArgs("quote-1").
		WillReturnRows(sqlmock.NewRows([]string{"quote_total_cents"}).AddRow(int64(1500)))

	resp, err := srv.GetSchedule(context.Background(), &scheduleGrpc.GetScheduleRequest{
		ScheduleId: "schedule-1",
		UserId:     "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.Schedule == nil {
		t.Fatal("expected schedule payload")
	}
	if resp.Schedule.StartMonth != "2026-06" {
		t.Fatalf("expected start_month 2026-06, got %q", resp.Schedule.StartMonth)
	}
	if len(resp.Schedule.Lines) != 2 {
		t.Fatalf("expected 2 line summaries, got %d", len(resp.Schedule.Lines))
	}
	if len(resp.Schedule.ColumnTotals) != 2 {
		t.Fatalf("expected 2 column totals, got %d", len(resp.Schedule.ColumnTotals))
	}
	if resp.Schedule.QuoteTotalCents != 1500 {
		t.Fatalf("expected quote total 1500, got %d", resp.Schedule.QuoteTotalCents)
	}
	if resp.Schedule.PlannedTotalCents != 1400 {
		t.Fatalf("expected planned total 1400, got %d", resp.Schedule.PlannedTotalCents)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetSchedule_InvalidInput(t *testing.T) {
	srv, mock := setupServer(t)

	resp, err := srv.GetSchedule(context.Background(), &scheduleGrpc.GetScheduleRequest{
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

func TestGetSchedule_NotFound(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT quote_id, status, name, start_month, duration_months FROM schedules WHERE schedule_id=\$1 AND user_id=\$2`).
		WithArgs("schedule-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"quote_id", "status", "name", "start_month", "duration_months"}))

	resp, err := srv.GetSchedule(context.Background(), &scheduleGrpc.GetScheduleRequest{
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