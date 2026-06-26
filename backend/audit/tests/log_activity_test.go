package tests

import (
	"context"
	"fmt"
	"testing"

	auditGrpc "project-devis-audit/services/grpc"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestLogActivity_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`INSERT INTO activity_logs`).
		WithArgs(nil, "GET", "/api/quotes", int32(42), nil, `{"ok":true}`, int32(200)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.LogActivity(context.Background(), &auditGrpc.LogActivityRequest{
		UserId:     "",
		Method:     "GET",
		Url:        "/api/quotes",
		DurationMs: 42,
		ReqBody:    "",
		RespBody:   `{"ok":true}`,
		RespStatus: 200,
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

func TestLogActivity_WithUserID(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`INSERT INTO activity_logs`).
		WithArgs("user-abc", "POST", "/api/quotes", int32(100), `{"name":"test"}`, `{"id":"1"}`, int32(201)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.LogActivity(context.Background(), &auditGrpc.LogActivityRequest{
		UserId:     "user-abc",
		Method:     "POST",
		Url:        "/api/quotes",
		DurationMs: 100,
		ReqBody:    `{"name":"test"}`,
		RespBody:   `{"id":"1"}`,
		RespStatus: 201,
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

func TestLogActivity_DBError(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`INSERT INTO activity_logs`).
		WillReturnError(fmt.Errorf("connection lost"))

	resp, err := srv.LogActivity(context.Background(), &auditGrpc.LogActivityRequest{
		Method:     "GET",
		Url:        "/api/quotes",
		DurationMs: 10,
		RespBody:   `{}`,
		RespStatus: 500,
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
