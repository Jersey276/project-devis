package tests

import (
	"context"
	"testing"
	"time"

	"project-devis-users/actions"
	usersGrpc "project-devis-users/services/grpc"

	"github.com/DATA-DOG/go-sqlmock"
)

// taxRows mirrors tax.Columns so mocks return the exact columns ScanRow expects.
func taxRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"id", "name", "rate", "country_group_id", "is_default",
		"original_tax_id", "version", "superseded_at", "superseded_by",
	})
}

// taxUpdatePreflightRows mirrors the SELECT in tax.Update's loadCurrent.
func taxUpdatePreflightRows() *sqlmock.Rows {
	return sqlmock.NewRows([]string{
		"name", "rate", "country_group_id", "original_tax_id", "superseded_at",
	})
}

func TestCreateTax_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO taxes`).
		WithArgs("TVA", "20.00", int32(1), false).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(1))
	mock.ExpectCommit()

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

func TestCreateTax_DefaultClearsSiblings(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectBegin()
	mock.ExpectExec(`UPDATE taxes SET is_default=FALSE`).
		WithArgs(int32(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectQuery(`INSERT INTO taxes`).
		WithArgs("TVA", "20.00", int32(1), true).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(2))
	mock.ExpectCommit()

	resp, err := srv.CreateTax(context.Background(), &usersGrpc.CreateTaxRequest{
		Name:           "TVA",
		Rate:           "20.00",
		CountryGroupId: 1,
		IsDefault:      true,
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
		WillReturnRows(taxRows().AddRow(1, "TVA", "20.00", 1, false, int32(0), int32(1), "", int32(0)))

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
	if resp.Tax.Version != 1 {
		t.Fatalf("expected version 1, got %d", resp.Tax.Version)
	}
}

func TestGetTax_NotFound(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT id, name, rate`).
		WithArgs(int32(999)).
		WillReturnRows(taxRows())

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

func TestDeleteTax_RetiresInPlace(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`UPDATE taxes SET superseded_at=NOW\(\), is_default=FALSE WHERE id=\$1 AND superseded_at IS NULL`).
		WithArgs(int32(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := srv.DeleteTax(context.Background(), &usersGrpc.DeleteTaxRequest{TaxId: 1})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
}

func TestUpdateTax_NameChange_CreatesNewVersion(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT name, rate::TEXT`).
		WithArgs(int32(1)).
		WillReturnRows(taxUpdatePreflightRows().AddRow("TVA", "20.00", 5, nil, nil))
	mock.ExpectQuery(`SELECT COALESCE\(MAX\(version\), 0\) \+ 1 FROM taxes`).
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"v"}).AddRow(2))
	mock.ExpectQuery(`INSERT INTO taxes`).
		WithArgs("TVA réduite", "20.00", int32(5), false, int32(1), int32(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(99))
	mock.ExpectExec(`UPDATE taxes SET superseded_at=NOW\(\), superseded_by=\$1, is_default=FALSE WHERE id=\$2`).
		WithArgs(int32(99), int32(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	resp, err := srv.UpdateTax(context.Background(), &usersGrpc.UpdateTaxRequest{
		TaxId: 1,
		Name:  "TVA réduite",
		Rate:  "20.00",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.TaxId != 99 {
		t.Fatalf("expected new tax_id 99, got %d", resp.TaxId)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestUpdateTax_RateChange_CreatesNewVersion(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT name, rate::TEXT`).
		WithArgs(int32(1)).
		WillReturnRows(taxUpdatePreflightRows().AddRow("TVA", "20.00", 5, nil, nil))
	mock.ExpectQuery(`SELECT COALESCE\(MAX\(version\), 0\) \+ 1`).
		WithArgs(int32(1)).
		WillReturnRows(sqlmock.NewRows([]string{"v"}).AddRow(2))
	mock.ExpectQuery(`INSERT INTO taxes`).
		WithArgs("TVA", "21.00", int32(5), false, int32(1), int32(2)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(99))
	mock.ExpectExec(`UPDATE taxes SET superseded_at=NOW\(\)`).
		WithArgs(int32(99), int32(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	resp, err := srv.UpdateTax(context.Background(), &usersGrpc.UpdateTaxRequest{
		TaxId: 1, Name: "TVA", Rate: "21.00",
	})
	if err != nil || !resp.Success {
		t.Fatalf("expected success, got resp=%+v err=%v", resp, err)
	}
	if resp.TaxId != 99 {
		t.Fatalf("expected new id 99, got %d", resp.TaxId)
	}
}

func TestUpdateTax_OnlyIsDefault_NoNewVersion(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT name, rate::TEXT`).
		WithArgs(int32(1)).
		WillReturnRows(taxUpdatePreflightRows().AddRow("TVA", "20.00", 5, nil, nil))
	mock.ExpectExec(`UPDATE taxes SET is_default=FALSE WHERE country_group_id=\$1 AND id<>\$2`).
		WithArgs(int32(5), int32(1)).
		WillReturnResult(sqlmock.NewResult(0, 0))
	mock.ExpectExec(`UPDATE taxes SET is_default=\$1 WHERE id=\$2`).
		WithArgs(true, int32(1)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	resp, err := srv.UpdateTax(context.Background(), &usersGrpc.UpdateTaxRequest{
		TaxId: 1, Name: "TVA", Rate: "20.00", IsDefault: true,
	})
	if err != nil || !resp.Success {
		t.Fatalf("expected success, got resp=%+v err=%v", resp, err)
	}
	if resp.TaxId != 1 {
		t.Fatalf("expected same id 1, got %d", resp.TaxId)
	}
}

func TestUpdateTax_NewVersionInheritsFamily(t *testing.T) {
	srv, mock := setupServer(t)

	// Updating v2 (id=99, original_tax_id=1) should inherit family root id=1.
	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT name, rate::TEXT`).
		WithArgs(int32(99)).
		WillReturnRows(taxUpdatePreflightRows().AddRow("TVA", "21.00", 5, 1, nil))
	mock.ExpectQuery(`SELECT COALESCE\(MAX\(version\), 0\) \+ 1 FROM taxes`).
		WithArgs(int32(1)). // family root, not the row id
		WillReturnRows(sqlmock.NewRows([]string{"v"}).AddRow(3))
	mock.ExpectQuery(`INSERT INTO taxes`).
		WithArgs("TVA", "22.00", int32(5), false, int32(1), int32(3)).
		WillReturnRows(sqlmock.NewRows([]string{"id"}).AddRow(150))
	mock.ExpectExec(`UPDATE taxes SET superseded_at=NOW\(\)`).
		WithArgs(int32(150), int32(99)).
		WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()

	resp, err := srv.UpdateTax(context.Background(), &usersGrpc.UpdateTaxRequest{
		TaxId: 99, Name: "TVA", Rate: "22.00",
	})
	if err != nil || !resp.Success {
		t.Fatalf("expected success, got resp=%+v err=%v", resp, err)
	}
	if resp.TaxId != 150 {
		t.Fatalf("expected id 150, got %d", resp.TaxId)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet: %v", err)
	}
}

func TestUpdateTax_SupersededRowIsImmutable(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`SELECT name, rate::TEXT`).
		WithArgs(int32(1)).
		WillReturnRows(taxUpdatePreflightRows().AddRow("TVA", "20.00", 5, nil, time.Now()))
	mock.ExpectRollback()

	resp, err := srv.UpdateTax(context.Background(), &usersGrpc.UpdateTaxRequest{
		TaxId: 1, Name: "TVA new", Rate: "20.00",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure when updating a superseded row")
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %d", resp.Code)
	}
}

func TestListTaxesForCountry_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectQuery(`SELECT id, name, rate.*FROM taxes.*country_group_countries.*country_id = \$1`).
		WithArgs(int32(42)).
		WillReturnRows(taxRows().
			AddRow(7, "USt 19%", "19.00", 3, true, int32(0), int32(1), "", int32(0)))

	resp, err := srv.ListTaxesForCountry(context.Background(), &usersGrpc.ListTaxesForCountryRequest{CountryId: 42})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if len(resp.Taxes) != 1 || resp.Taxes[0].Rate != "19.00" {
		t.Fatalf("expected one tax at 19.00, got %+v", resp.Taxes)
	}
}

func TestListTaxesForCountry_MissingCountryID(t *testing.T) {
	srv, _ := setupServer(t)

	resp, err := srv.ListTaxesForCountry(context.Background(), &usersGrpc.ListTaxesForCountryRequest{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for missing country_id")
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %d", resp.Code)
	}
}
