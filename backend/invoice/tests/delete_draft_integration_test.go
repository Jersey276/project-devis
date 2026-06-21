package tests

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"project-devis-invoice/actions"
	"project-devis-invoice/actions/codes"
	invoiceGrpc "project-devis-invoice/services/grpc"

	_ "github.com/lib/pq"
)

func seedDraftInvoice(t *testing.T, db *sql.DB, userID, invoiceID string) {
	t.Helper()
	if _, err := db.Exec(
		`INSERT INTO invoices (invoice_id, user_id, quote_id, status)
		 VALUES ($1, $2, 'quote-x', 'DRAFT')`,
		invoiceID, userID,
	); err != nil {
		t.Fatalf("seed draft: %v", err)
	}
}

func invoiceExists(t *testing.T, db *sql.DB, invoiceID string) bool {
	t.Helper()
	var exists bool
	if err := db.QueryRow(
		`SELECT EXISTS(SELECT 1 FROM invoices WHERE invoice_id=$1)`, invoiceID,
	).Scan(&exists); err != nil {
		t.Fatalf("exists check: %v", err)
	}
	return exists
}

func TestDeleteDraft_RemovesDraft(t *testing.T) {
	db := sealTestDB(t)
	const userID = "del-test-draft"
	seedDraftInvoice(t, db, userID, "inv-draft")

	srv := actions.NewServer(db, nil, nil, nil, nil, nil)
	resp, err := srv.DeleteDraftInvoice(context.Background(), &invoiceGrpc.DeleteDraftInvoiceRequest{
		InvoiceId: "inv-draft", UserId: userID,
	})
	if err != nil {
		t.Fatalf("delete draft: %v", err)
	}
	if !resp.Success || resp.Code != codes.Success {
		t.Fatalf("delete draft: success=%v code=%d; want success/0", resp.Success, resp.Code)
	}
	if invoiceExists(t, db, "inv-draft") {
		t.Fatal("draft invoice still present after delete")
	}
}

func TestDeleteDraft_RefusesIssued(t *testing.T) {
	db := sealTestDB(t)
	const userID = "del-test-issued"

	seedIssuedInvoice(t, db, userID, "inv-issued", "2099-0001",
		time.Date(2099, 4, 1, 9, 0, 0, 0, time.UTC), 1)
	if err := actions.BackfillSeals(context.Background(), db); err != nil {
		t.Fatalf("backfill: %v", err)
	}

	srv := actions.NewServer(db, nil, nil, nil, nil, nil)
	resp, err := srv.DeleteDraftInvoice(context.Background(), &invoiceGrpc.DeleteDraftInvoiceRequest{
		InvoiceId: "inv-issued", UserId: userID,
	})
	if err != nil {
		t.Fatalf("delete issued: unexpected error: %v", err)
	}
	if resp.Success {
		t.Fatal("delete on an issued invoice succeeded; it must be refused")
	}
	if resp.Code != codes.InvoiceFinalized {
		t.Fatalf("delete issued code=%d; want InvoiceFinalized (%d)", resp.Code, codes.InvoiceFinalized)
	}
	if !invoiceExists(t, db, "inv-issued") {
		t.Fatal("issued invoice was deleted; it must stay immutable")
	}
}

func TestDeleteDraft_NotFound(t *testing.T) {
	db := sealTestDB(t)
	srv := actions.NewServer(db, nil, nil, nil, nil, nil)
	resp, err := srv.DeleteDraftInvoice(context.Background(), &invoiceGrpc.DeleteDraftInvoiceRequest{
		InvoiceId: "does-not-exist", UserId: "del-test-missing",
	})
	if err != nil {
		t.Fatalf("delete missing: %v", err)
	}
	if resp.Success || resp.Code != codes.NotFound {
		t.Fatalf("delete missing: success=%v code=%d; want NotFound (%d)", resp.Success, resp.Code, codes.NotFound)
	}
}
