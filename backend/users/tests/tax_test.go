package tests

import (
	"context"
	"testing"

	"project-devis-users/actions"
	usersGrpc "project-devis-users/services/grpc"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestCreateTax_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`INSERT INTO taxes`).
		WithArgs("TVA", "20.00", int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))

	resp, err := srv.CreateTax(context.Background(), &usersGrpc.CreateTaxRequest{
		Name:           "TVA",
		Rate:           "20.00",
		CountryGroupId: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.TaxId != 1 {
		t.Fatalf("expected tax_id 1, got %d", resp.TaxId)
	}
}

func TestCreateTax_InvalidRate(t *testing.T) {
	srv, mock := setupServer(t)

	resp, err := srv.CreateTax(context.Background(), &usersGrpc.CreateTaxRequest{
		Name:           "TVA",
		Rate:           "not-a-number",
		CountryGroupId: 1,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for invalid rate")
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB calls: %v", err)
	}
}

func TestCreateTax_MissingFields(t *testing.T) {
	srv, mock := setupServer(t)

	resp, err := srv.CreateTax(context.Background(), &usersGrpc.CreateTaxRequest{Name: "TVA"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for missing fields")
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB calls: %v", err)
	}
}

func TestGetTax_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT id, name, rate`).
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "rate", "country_group_id"}).
			AddRow(1, "TVA", "20.00", 1))

	resp, err := srv.GetTax(context.Background(), &usersGrpc.GetTaxRequest{TaxId: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.Tax.Name != "TVA" {
		t.Fatalf("expected name TVA, got %q", resp.Tax.Name)
	}
	if resp.Tax.Rate != "20.00" {
		t.Fatalf("expected rate 20.00, got %q", resp.Tax.Rate)
	}
}

func TestGetTax_NotFound(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT id, name, rate`).
		WithArgs(int32(999)).
		WillReturnRows(sqlmock.NewRows([]string{"id", "name", "rate", "country_group_id"}))

	resp, err := srv.GetTax(context.Background(), &usersGrpc.GetTaxRequest{TaxId: 999})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for non-existent tax")
	}
	if resp.Code != actions.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %d", resp.Code)
	}
}

func TestDeleteTax_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`DELETE FROM taxes`).
		WithArgs(int32(1)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.DeleteTax(context.Background(), &usersGrpc.DeleteTaxRequest{TaxId: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
}
