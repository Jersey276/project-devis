package tests

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"project-devis-invoice/actions"
	invoiceGrpc "project-devis-invoice/services/grpc"

	_ "github.com/lib/pq"
)

// Same disposable-schema harness as the seal tests (sealTestDB / seedIssuedInvoice).

// seedPartySnapshot writes a party snapshot row directly, including the buyer
// SIREN/VAT (migration 000004), client_type (000005), client_country_id
// (000006) and oss_applied (000007).
func seedPartySnapshot(t *testing.T, db *sql.DB, invoiceID, clientSiren, clientVat, clientType string, clientCountryID int32, ossApplied bool) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO invoice_party_snapshots (
			invoice_id,
			issuer_company, issuer_siren, issuer_vat,
			client_first_name, client_last_name, client_company, client_siren, client_vat,
			client_type, client_country_id, oss_applied
		) VALUES ($1, 'Acme SARL', '123456782', 'FR12345678901',
		          'Jean', 'Dupont', 'Dupont SAS', $2, $3, $4, $5, $6)`,
		invoiceID, clientSiren, clientVat, clientType, clientCountryID, ossApplied)
	if err != nil {
		t.Fatalf("seed party snapshot: %v", err)
	}
}

// TestPartySnapshot_ExposesClientTaxIds verifies the round-trip added in step 1:
// client_siren/client_vat are stored, read back and surfaced on InvoiceParty so
// the Factur-X generator can map the buyer's tax identifiers.
func TestPartySnapshot_ExposesClientTaxIds(t *testing.T) {
	db := sealTestDB(t)
	const userID = "party-test"

	seedIssuedInvoice(t, db, userID, "inv-party", "2099-0001",
		time.Date(2099, 5, 1, 9, 0, 0, 0, time.UTC), 1)
	seedPartySnapshot(t, db, "inv-party", "987654321", "FR99887766554", "business", 42, false)

	srv := actions.NewServer(db, nil, nil, nil)
	resp, err := srv.GetInvoice(context.Background(), &invoiceGrpc.GetInvoiceRequest{
		InvoiceId: "inv-party", UserId: userID,
	})
	if err != nil {
		t.Fatalf("get invoice: %v", err)
	}
	if !resp.Success || resp.Invoice == nil {
		t.Fatalf("get invoice: success=%v invoice=%v code=%d", resp.Success, resp.Invoice, resp.Code)
	}

	client := resp.Invoice.GetClient()
	if got := client.GetSiren(); got != "987654321" {
		t.Errorf("client SIREN = %q; want 987654321", got)
	}
	if got := client.GetVat(); got != "FR99887766554" {
		t.Errorf("client VAT = %q; want FR99887766554", got)
	}
	if got := client.GetClientType(); got != "business" {
		t.Errorf("client type = %q; want business", got)
	}
	if got := client.GetClientCountryId(); got != 42 {
		t.Errorf("client country id = %d; want 42", got)
	}

	// Issuer side must still carry its own tax ids (regression guard).
	if got := resp.Invoice.GetIssuer().GetSiren(); got != "123456782" {
		t.Errorf("issuer SIREN = %q; want 123456782", got)
	}
}

// TestPartySnapshot_EmptyClientTaxIds covers a legacy/B2C row: missing buyer tax
// ids come back as empty strings (default ''), never a scan error.
func TestPartySnapshot_EmptyClientTaxIds(t *testing.T) {
	db := sealTestDB(t)
	const userID = "party-test-empty"

	seedIssuedInvoice(t, db, userID, "inv-empty", "2099-0002",
		time.Date(2099, 5, 2, 9, 0, 0, 0, time.UTC), 2)
	seedPartySnapshot(t, db, "inv-empty", "", "", "", 0, false)

	srv := actions.NewServer(db, nil, nil, nil)
	resp, err := srv.GetInvoice(context.Background(), &invoiceGrpc.GetInvoiceRequest{
		InvoiceId: "inv-empty", UserId: userID,
	})
	if err != nil {
		t.Fatalf("get invoice: %v", err)
	}
	if !resp.Success || resp.Invoice == nil {
		t.Fatalf("get invoice failed: code=%d", resp.Code)
	}
	if got := resp.Invoice.GetClient().GetSiren(); got != "" {
		t.Errorf("client SIREN = %q; want empty", got)
	}
}

// TestPartySnapshot_OssApplied verifies the OSS flag (migration 000007) is
// stored and surfaced on InvoiceDetails, so the export PDF can print the
// guichet-unique mention from the snapshot alone.
func TestPartySnapshot_OssApplied(t *testing.T) {
	db := sealTestDB(t)
	const userID = "party-test-oss"

	seedIssuedInvoice(t, db, userID, "inv-oss", "2099-0003",
		time.Date(2099, 5, 3, 9, 0, 0, 0, time.UTC), 3)
	seedPartySnapshot(t, db, "inv-oss", "", "", "individual", 276, true)

	srv := actions.NewServer(db, nil, nil, nil)
	resp, err := srv.GetInvoice(context.Background(), &invoiceGrpc.GetInvoiceRequest{
		InvoiceId: "inv-oss", UserId: userID,
	})
	if err != nil {
		t.Fatalf("get invoice: %v", err)
	}
	if !resp.Success || resp.Invoice == nil {
		t.Fatalf("get invoice failed: code=%d", resp.Code)
	}
	if !resp.Invoice.GetOssApplied() {
		t.Errorf("oss_applied = false; want true")
	}
}
