package tests

import (
	"context"
	"testing"

	templateGrpc "project-devis-template/services/grpc"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCreateLine_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT COUNT\(1\) FROM templates`).
		WithArgs("tmpl-1", "user-1").
		WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

	mock.ExpectExec(`INSERT INTO template_lines`).
		WithArgs(sqlmock.AnyArg(), "tmpl-1", "simple", "Prestation", "1", nil, int64(10000), "{}", int32(0), nil).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.CreateTemplateLine(context.Background(), &templateGrpc.CreateTemplateLineRequest{
		TemplateId: "tmpl-1",
		UserId:     "user-1",
		Type:       "simple",
		Name:       "Prestation",
		Quantity:   "1",
		UnitPrice:  10000,
		Position:   0,
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

func TestCreateLine_InvalidQuantity(t *testing.T) {
	srv, mock := setupServer(t)

	resp, err := srv.CreateTemplateLine(context.Background(), &templateGrpc.CreateTemplateLineRequest{
		TemplateId: "tmpl-1",
		UserId:     "user-1",
		Type:       "simple",
		Name:       "Test",
		Quantity:   "not-a-number",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for invalid quantity")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB calls: %v", err)
	}
}

func TestDeleteLine_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`DELETE FROM template_lines`).
		WithArgs("line-1", "tmpl-1", "user-1").
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.DeleteTemplateLine(context.Background(), &templateGrpc.DeleteTemplateLineRequest{
		LineId:     "line-1",
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
