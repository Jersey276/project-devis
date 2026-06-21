package tests

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"project-devis-invoice/actions"
	"project-devis-invoice/actions/codes"
	"project-devis-invoice/pdp"
	invoiceGrpc "project-devis-invoice/services/grpc"

	_ "github.com/lib/pq"
)

// seedRecipientSnapshot inserts a minimal legal snapshot carrying the recipient
// SIRET, so the deposit flow can read client_siret for the directory lookup.
func seedRecipientSnapshot(t *testing.T, db *sql.DB, invoiceID, clientSiret string) {
	t.Helper()
	if _, err := db.Exec(
		`INSERT INTO invoice_party_snapshots (invoice_id, client_siret) VALUES ($1, $2)`,
		invoiceID, clientSiret,
	); err != nil {
		t.Fatalf("seed party snapshot: %v", err)
	}
}

func recipientRoutingID(t *testing.T, db *sql.DB, invoiceID string) string {
	t.Helper()
	var rid sql.NullString
	if err := db.QueryRow(
		`SELECT recipient_routing_id FROM invoices WHERE invoice_id=$1`, invoiceID,
	).Scan(&rid); err != nil {
		t.Fatalf("read recipient_routing_id: %v", err)
	}
	return rid.String
}

func TestDeposit_ResolvesRecipientAndFreezesRouting(t *testing.T) {
	db := sealTestDB(t)
	const userID = "deposit-dir-ok"
	seedIssuedInvoice(t, db, userID, "inv-dir", "2099-0001",
		time.Date(2099, 4, 1, 9, 0, 0, 0, time.UTC), 1)
	seedRecipientSnapshot(t, db, "inv-dir", "12345678900011")

	mock := &pdp.MockClient{SubmitResult: pdp.SubmitResult{SubmissionID: "sub-1", Status: pdp.PlatformSubmitted}}
	dir := &pdp.MockDirectory{ResolveResult: pdp.RecipientRouting{RoutingID: "route-1", PlatformName: "PA Test"}}
	srv := actions.NewServer(db, nil, nil, nil, mock, dir)

	resp, err := srv.DepositInvoice(context.Background(), &invoiceGrpc.DepositInvoiceRequest{InvoiceId: "inv-dir", UserId: userID})
	if err != nil {
		t.Fatalf("deposit: %v", err)
	}
	if !resp.Success || resp.Code != codes.Success {
		t.Fatalf("deposit: success=%v code=%d; want success/0", resp.Success, resp.Code)
	}
	// Directory was queried with the frozen recipient SIRET.
	if len(dir.Resolved) != 1 || dir.Resolved[0] != "12345678900011" {
		t.Fatalf("directory Resolve calls=%+v; want one with SIRET 12345678900011", dir.Resolved)
	}
	if got := recipientRoutingID(t, db, "inv-dir"); got != "route-1" {
		t.Fatalf("recipient_routing_id=%q; want route-1", got)
	}
	if got := lifecycleState(t, db, "inv-dir"); got != "DEPOSITED" {
		t.Fatalf("lifecycle_status=%q; want DEPOSITED", got)
	}
}

func TestDeposit_RecipientNotInDirectoryBlocks(t *testing.T) {
	db := sealTestDB(t)
	const userID = "deposit-dir-missing"
	seedIssuedInvoice(t, db, userID, "inv-dir-miss", "2099-0001",
		time.Date(2099, 4, 1, 9, 0, 0, 0, time.UTC), 1)
	seedRecipientSnapshot(t, db, "inv-dir-miss", "99999999900099")

	mock := &pdp.MockClient{SubmitResult: pdp.SubmitResult{SubmissionID: "sub-1", Status: pdp.PlatformSubmitted}}
	dir := &pdp.MockDirectory{ResolveErr: pdp.ErrRecipientNotFound}
	srv := actions.NewServer(db, nil, nil, nil, mock, dir)

	resp, err := srv.DepositInvoice(context.Background(), &invoiceGrpc.DepositInvoiceRequest{InvoiceId: "inv-dir-miss", UserId: userID})
	if err != nil {
		t.Fatalf("deposit: %v", err)
	}
	if resp.Success || resp.Code != codes.RecipientNotInDirectory {
		t.Fatalf("success=%v code=%d; want RecipientNotInDirectory (%d)", resp.Success, resp.Code, codes.RecipientNotInDirectory)
	}
	// An unresolved recipient must never reach the platform, and state stays put.
	if len(mock.Submitted) != 0 {
		t.Fatalf("platform was called despite unresolved recipient: %+v", mock.Submitted)
	}
	if got := lifecycleState(t, db, "inv-dir-miss"); got != "NONE" {
		t.Fatalf("lifecycle_status=%q; want NONE (untouched)", got)
	}
	if got := recipientRoutingID(t, db, "inv-dir-miss"); got != "" {
		t.Fatalf("recipient_routing_id=%q; want empty", got)
	}
}

func TestDeposit_NoopDirectoryDepositsWithoutRouting(t *testing.T) {
	db := sealTestDB(t)
	const userID = "deposit-dir-noop"
	seedIssuedInvoice(t, db, userID, "inv-dir-noop", "2099-0001",
		time.Date(2099, 4, 1, 9, 0, 0, 0, time.UTC), 1)
	seedRecipientSnapshot(t, db, "inv-dir-noop", "12345678900011")

	// nil directory => NoopDirectory: resolves everyone, empty routing handle.
	mock := &pdp.MockClient{SubmitResult: pdp.SubmitResult{SubmissionID: "sub-1", Status: pdp.PlatformSubmitted}}
	srv := actions.NewServer(db, nil, nil, nil, mock, nil)

	resp, err := srv.DepositInvoice(context.Background(), &invoiceGrpc.DepositInvoiceRequest{InvoiceId: "inv-dir-noop", UserId: userID})
	if err != nil {
		t.Fatalf("deposit: %v", err)
	}
	if !resp.Success || resp.Code != codes.Success {
		t.Fatalf("deposit: success=%v code=%d; want success/0", resp.Success, resp.Code)
	}
	if got := lifecycleState(t, db, "inv-dir-noop"); got != "DEPOSITED" {
		t.Fatalf("lifecycle_status=%q; want DEPOSITED", got)
	}
	// No real annuaire: routing handle stays NULL/empty.
	if got := recipientRoutingID(t, db, "inv-dir-noop"); got != "" {
		t.Fatalf("recipient_routing_id=%q; want empty (no-op directory)", got)
	}
}
