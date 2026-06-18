package tests

import (
	"context"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"project-devis-quote/actions"
	quoteGrpc "project-devis-quote/services/grpc"
)

func TestCreateFee_Success(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`INSERT INTO fees`).
		WithArgs(sqlmock.AnyArg(), "user-1", "service", "Livraison", "h", int64(5000), int32(3)).
		WillReturnResult(sqlmock.NewResult(1, 1))

	resp, err := srv.CreateFee(context.Background(), &quoteGrpc.CreateFeeRequest{
		UserId:    "user-1",
		Category:  "service",
		Name:      "Livraison",
		Unit:      "h",
		UnitPrice: 5000,
		TaxId:     3,
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("expected success, got code %d", resp.Code)
	}
	if resp.FeeId == "" {
		t.Fatal("expected non-empty fee_id")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestCreateFee_InvalidCategory(t *testing.T) {
	srv, mock := setupServer(t)

	resp, err := srv.CreateFee(context.Background(), &quoteGrpc.CreateFeeRequest{
		UserId:   "user-1",
		Category: "weird",
		Name:     "x",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for invalid category")
	}
	if resp.Code != actions.CodeInvalidInput {
		t.Fatalf("expected CodeInvalidInput, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB calls: %v", err)
	}
}

// TestUpdateFee_PropagatesToEditableQuotes verifies that updating a fee writes
// the new snapshot AND fans out the propagation queries. The propagation is
// restricted to draft/sent quotes by the SQL itself (state = ANY(...)), so a
// validated quote can never be touched — this is asserted by matching the
// editable-states clause in both propagation queries.
func TestUpdateFee_PropagatesToEditableQuotes(t *testing.T) {
	srv, mock := setupServer(t)

	// 1. The fee row is updated.
	mock.ExpectExec(`UPDATE fees`).
		WithArgs("service", "Livraison", "h", int64(6000), int32(3), "fee-1", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 1))

	// 2. Top-level fee lines refreshed (name/unit/price only — tax_id is not
	//    propagated), scoped to draft/sent quotes only.
	mock.ExpectExec(`UPDATE quote_lines l\s+SET name=\$1, unit=\$2, unit_price=\$3, updated_at=NOW\(\) FROM quotes q\s+WHERE .* l\.fee_id = \$4 .* q\.state = ANY\(\$6\)`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	// 3. Detailed lines with fee sublines are read (same editable-states scope)...
	mock.ExpectQuery(`SELECT l\.line_id, l\.data::text\s+FROM quote_lines l\s+JOIN quotes q .* q\.state = ANY\(\$2\) .* l\.data @> \$3::jsonb`).
		WillReturnRows(sqlmock.NewRows([]string{"line_id", "data"}).
			AddRow("line-1", `{"kind":"detailed","sublines":[{"name":"old","quantity":"1","unit_price":1,"fee_id":"fee-1"}]}`))

	// 4. ...and all matching sublines are rewritten in a single batched UPDATE.
	mock.ExpectExec(`UPDATE quote_lines l\s+SET data = p\.data::jsonb.* FROM \(SELECT unnest\(\$1::text\[\]\) AS line_id, unnest\(\$2::text\[\]\) AS data\) p\s+WHERE l\.line_id = p\.line_id`).
		WillReturnResult(sqlmock.NewResult(0, 1))

	resp, err := srv.UpdateFee(context.Background(), &quoteGrpc.UpdateFeeRequest{
		FeeId:     "fee-1",
		UserId:    "user-1",
		Category:  "service",
		Name:      "Livraison",
		Unit:      "h",
		UnitPrice: 6000,
		TaxId:     3,
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

func TestUpdateFee_NotFound(t *testing.T) {
	srv, mock := setupServer(t)

	mock.ExpectExec(`UPDATE fees`).
		WithArgs("fixed", "x", nil, int64(0), nil, "missing", "user-1").
		WillReturnResult(sqlmock.NewResult(0, 0))

	resp, err := srv.UpdateFee(context.Background(), &quoteGrpc.UpdateFeeRequest{
		FeeId:    "missing",
		UserId:   "user-1",
		Category: "fixed",
		Name:     "x",
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("expected failure for missing fee")
	}
	if resp.Code != actions.CodeNotFound {
		t.Fatalf("expected CodeNotFound, got %d", resp.Code)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
