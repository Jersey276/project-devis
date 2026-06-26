package tests

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	auditGrpc "project-devis-audit/services/grpc"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetActivityLog_Success(t *testing.T) {
	srv, mock := setupServer(t)

	rows := sqlmock.NewRows([]string{
		"id", "user_id", "method", "url", "duration_ms",
		"req_body", "resp_body", "resp_status", "created_at",
	}).AddRow(int64(1), "user-1", "GET", "/api/quotes", int32(55), "", `{"data":[]}`, int32(200), "2024-01-15T10:00:00Z")

	mock.ExpectQuery(`SELECT id`).
		WithArgs(int64(1)).
		WillReturnRows(rows)

	resp, err := srv.GetActivityLog(context.Background(), &auditGrpc.GetActivityLogRequest{Id: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.Log == nil {
		t.Fatal("expected log in response")
	}
	if resp.Log.Method != "GET" {
		t.Fatalf("expected method GET, got %s", resp.Log.Method)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetActivityLog_NotFound(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT id`).
		WithArgs(int64(999)).
		WillReturnError(sql.ErrNoRows)

	resp, err := srv.GetActivityLog(context.Background(), &auditGrpc.GetActivityLogRequest{Id: 999})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure")
	}
	if resp.Code != 3 {
		t.Fatalf("expected CodeNotFound=3, got %d", resp.Code)
	}
}

func TestGetActivityLog_DBError(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT id`).
		WithArgs(int64(1)).
		WillReturnError(fmt.Errorf("db timeout"))

	resp, err := srv.GetActivityLog(context.Background(), &auditGrpc.GetActivityLogRequest{Id: 1})
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
