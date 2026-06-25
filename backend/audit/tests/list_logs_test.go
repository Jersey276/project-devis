package tests

import (
	"context"
	"fmt"
	"testing"

	auditGrpc "project-devis-audit/services/grpc"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestListActivityLogs_NoFilters(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM activity_logs`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(2)))

	rows := sqlmock.NewRows([]string{"id", "user_id", "method", "url", "duration_ms", "resp_status", "created_at"}).
		AddRow(int64(1), "u1", "GET", "/api/a", int32(10), int32(200), "2024-01-01T00:00:00Z").
		AddRow(int64(2), "u2", "POST", "/api/b", int32(20), int32(201), "2024-01-02T00:00:00Z")

	mock.ExpectQuery(`SELECT id`).
		WillReturnRows(rows)

	resp, err := srv.ListActivityLogs(context.Background(), &auditGrpc.ListActivityLogsRequest{
		Page:     1,
		PageSize: 50,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.Total != 2 {
		t.Fatalf("expected total 2, got %d", resp.Total)
	}
	if len(resp.Logs) != 2 {
		t.Fatalf("expected 2 logs, got %d", len(resp.Logs))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestListActivityLogs_FilterByUserID(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM activity_logs`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(1)))

	rows := sqlmock.NewRows([]string{"id", "user_id", "method", "url", "duration_ms", "resp_status", "created_at"}).
		AddRow(int64(1), "user-x", "GET", "/api/quotes", int32(5), int32(200), "2024-01-01T00:00:00Z")

	mock.ExpectQuery(`SELECT id`).
		WillReturnRows(rows)

	resp, err := srv.ListActivityLogs(context.Background(), &auditGrpc.ListActivityLogsRequest{
		Filters:  &auditGrpc.ActivityLogFilters{UserId: "user-x"},
		Page:     1,
		PageSize: 20,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.Total != 1 {
		t.Fatalf("expected total 1, got %d", resp.Total)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestListActivityLogs_FilterByStatus(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM activity_logs`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	mock.ExpectQuery(`SELECT id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "method", "url", "duration_ms", "resp_status", "created_at"}))

	resp, err := srv.ListActivityLogs(context.Background(), &auditGrpc.ListActivityLogsRequest{
		Filters:  &auditGrpc.ActivityLogFilters{RespStatuses: []int32{500, 502}},
		Page:     1,
		PageSize: 50,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.Total != 0 {
		t.Fatalf("expected total 0, got %d", resp.Total)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestListActivityLogs_PageSizeClamped(t *testing.T) {
	srv, mock := setupServer(t)

	// pageSize=500 doit être clampé à 50 — on vérifie que la query s'exécute sans erreur
	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM activity_logs`).
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(int64(0)))

	mock.ExpectQuery(`SELECT id`).
		WillReturnRows(sqlmock.NewRows([]string{"id", "user_id", "method", "url", "duration_ms", "resp_status", "created_at"}))

	resp, err := srv.ListActivityLogs(context.Background(), &auditGrpc.ListActivityLogsRequest{
		Page:     1,
		PageSize: 500,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
}

func TestListActivityLogs_DBError(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT COUNT\(\*\) FROM activity_logs`).
		WillReturnError(fmt.Errorf("connection refused"))

	resp, err := srv.ListActivityLogs(context.Background(), &auditGrpc.ListActivityLogsRequest{
		Page:     1,
		PageSize: 50,
	})
	if err != nil {
		t.Fatalf("unexpected error returned: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure")
	}
	if resp.Code != 1 {
		t.Fatalf("expected CodeInternalError=1, got %d", resp.Code)
	}
}
