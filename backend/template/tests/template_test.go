package tests

import (
	"context"
	"testing"

	templateGrpc "project-devis-template/services/grpc"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCreateTemplate_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`INSERT INTO templates`).
		WithArgs(sqlmock.AnyArg(), "user-1", "quote_document", "quote", "Mon Template").
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.CreateTemplate(context.Background(), &templateGrpc.CreateTemplateRequest{
		UserId:         "user-1",
		TemplateType:   "quote_document",
		TargetResource: "quote",
		Name:           "Mon Template",
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

func TestCreateTemplate_InvalidType(t *testing.T) {
	srv, mock := setupServer(t)

	resp, err := srv.CreateTemplate(context.Background(), &templateGrpc.CreateTemplateRequest{
		UserId:         "user-1",
		TemplateType:   "invalid_type",
		TargetResource: "quote",
		Name:           "Test",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for invalid template type")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB calls: %v", err)
	}
}

func TestCreateTemplate_MissingUserID(t *testing.T) {
	srv, mock := setupServer(t)

	resp, err := srv.CreateTemplate(context.Background(), &templateGrpc.CreateTemplateRequest{
		TemplateType:   "quote_document",
		TargetResource: "quote",
		Name:           "Test",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for missing user_id")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB calls: %v", err)
	}
}

func TestGetTemplate_NotFound(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT template_id`).
		WithArgs("unknown-id", "user-1").
		WillReturnRows(sqlmock.NewRows(nil))

	resp, err := srv.GetTemplate(context.Background(), &templateGrpc.GetTemplateRequest{
		TemplateId: "unknown-id",
		UserId:     "user-1",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for unknown template")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestDeleteTemplate_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`DELETE FROM templates`).
		WithArgs("tmpl-1", "user-1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.DeleteTemplate(context.Background(), &templateGrpc.DeleteTemplateRequest{
		TemplateId: "tmpl-1",
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

func TestArchiveTemplate_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`UPDATE templates SET archived_at`).
		WithArgs("tmpl-1", "user-1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.ArchiveTemplate(context.Background(), &templateGrpc.ArchiveTemplateRequest{
		TemplateId: "tmpl-1",
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
