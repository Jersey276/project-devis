package tests

import (
	"context"
	"fmt"
	"testing"

	auditGrpc "project-devis-audit/services/grpc"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestGetActivityStats_Success(t *testing.T) {
	srv, mock := setupServer(t)

	rows := sqlmock.NewRows([]string{"day", "resp_status", "count"}).
		AddRow("2024-01-01", int32(200), int64(10)).
		AddRow("2024-01-01", int32(500), int64(2))

	mock.ExpectQuery(`SELECT to_char`).
		WillReturnRows(rows)

	resp, err := srv.GetActivityStats(context.Background(), &auditGrpc.GetActivityStatsRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if len(resp.Stats) != 2 {
		t.Fatalf("expected 2 stat rows, got %d", len(resp.Stats))
	}
	if resp.Stats[0].Date != "2024-01-01" {
		t.Fatalf("expected date 2024-01-01, got %s", resp.Stats[0].Date)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestGetActivityStats_Empty(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT to_char`).
		WillReturnRows(sqlmock.NewRows([]string{"day", "resp_status", "count"}))

	resp, err := srv.GetActivityStats(context.Background(), &auditGrpc.GetActivityStatsRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if len(resp.Stats) != 0 {
		t.Fatalf("expected 0 stats, got %d", len(resp.Stats))
	}
}

func TestGetActivityStats_DBError(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT to_char`).
		WillReturnError(fmt.Errorf("timeout"))

	resp, err := srv.GetActivityStats(context.Background(), &auditGrpc.GetActivityStatsRequest{})
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
