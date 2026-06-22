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

// seedReportInvoice inserts an issued invoice with its frozen party snapshot
// (B2C) and one VAT-breakdown row, in the given destination country. counts sets
// the OSS assiette flag (only meaningful for the cross-border scope).
func seedReportInvoice(t *testing.T, db *sql.DB, userID, invoiceID, clientCountry string, issuedAt time.Time, htCents, vatCents int64, counts bool) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO invoices (invoice_id, user_id, quote_id, status, issued_at,
		                       total_ht_cents, total_vat_cents, total_ttc_cents, vat_exempt)
		 VALUES ($1,$2,'quote-x','ISSUED',$3,$4,$5,$4::bigint+$5::bigint,false)`,
		invoiceID, userID, issuedAt, htCents, vatCents)
	if err != nil {
		t.Fatalf("seed report invoice %s: %v", invoiceID, err)
	}
	seedPartySnapshot(t, db, invoiceID, "", "", "individual", 276, counts, "FR", clientCountry, counts)
	_, err = db.Exec(
		`INSERT INTO invoice_vat_breakdown_snapshots (invoice_id, tax_rate, base_ht_cents, vat_cents)
		 VALUES ($1,'20',$2,$3)`,
		invoiceID, htCents, vatCents)
	if err != nil {
		t.Fatalf("seed vat breakdown %s: %v", invoiceID, err)
	}
}

// seedReportCreditNote inserts a credit note with its frozen party + VAT
// snapshots, neutralising part of an in-scope invoice in the same period/country.
func seedReportCreditNote(t *testing.T, db *sql.DB, userID, creditNoteID, invoiceID, clientCountry string, issuedAt time.Time, htCents, vatCents int64, counts bool) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO credit_notes (credit_note_id, user_id, invoice_id, credit_note_number,
		                           number_year, number_seq, issued_at,
		                           total_ht_cents, total_vat_cents, total_ttc_cents, vat_exempt)
		 VALUES ($1,$2,$3,$4,$5,1,$6,$7,$8,$7::bigint+$8::bigint,false)`,
		creditNoteID, userID, invoiceID, "AV-"+creditNoteID, issuedAt.Year(), issuedAt, htCents, vatCents)
	if err != nil {
		t.Fatalf("seed report credit note %s: %v", creditNoteID, err)
	}
	_, err = db.Exec(
		`INSERT INTO credit_note_party_snapshots (credit_note_id, client_type, client_country_id,
		                                          oss_applied, issuer_country_code, client_country_code,
		                                          counts_toward_oss_threshold)
		 VALUES ($1,'individual',276,$2,'FR',$3,$2)`,
		creditNoteID, counts, clientCountry)
	if err != nil {
		t.Fatalf("seed report credit note party %s: %v", creditNoteID, err)
	}
	_, err = db.Exec(
		`INSERT INTO credit_note_vat_breakdown_snapshots (credit_note_id, tax_rate, base_ht_cents, vat_cents)
		 VALUES ($1,'20',$2,$3)`,
		creditNoteID, htCents, vatCents)
	if err != nil {
		t.Fatalf("seed report credit note vat %s: %v", creditNoteID, err)
	}
}

func reportRow(t *testing.T, db *sql.DB, userID string, kind pdp.ReportKind, year, month int) (status string, totalHT, totalVAT int64, reportID sql.NullString) {
	t.Helper()
	err := db.QueryRow(
		`SELECT status, total_ht_cents, total_vat_cents, report_id FROM invoice_reports
		  WHERE user_id=$1 AND kind=$2 AND period_year=$3 AND period_month=$4`,
		userID, string(kind), year, month,
	).Scan(&status, &totalHT, &totalVAT, &reportID)
	if err != nil {
		t.Fatalf("read report row: %v", err)
	}
	return
}

// INV-BE-090: a TRANSACTION report aggregates domestic B2C only, excluding
// intra-EU sales (which belong to the cross-border report), and records DEPOSITED.
func TestSubmitReport_TransactionScope(t *testing.T) {
	db := sealTestDB(t)
	const userID = "rep-tx"
	apr := time.Date(2099, 4, 10, 9, 0, 0, 0, time.UTC)

	seedReportInvoice(t, db, userID, "inv-fr-1", "FR", apr, 100_000, 20_000, false)
	seedReportInvoice(t, db, userID, "inv-fr-2", "FR", apr, 50_000, 10_000, false)
	// Intra-EU sale: must NOT land in the transaction report.
	seedReportInvoice(t, db, userID, "inv-de", "DE", apr, 999_999, 0, true)
	// Different month: excluded.
	seedReportInvoice(t, db, userID, "inv-mar", "FR", time.Date(2099, 3, 1, 9, 0, 0, 0, time.UTC), 70_000, 14_000, false)

	mock := &pdp.MockReporter{SubmitResult: pdp.SubmitReportResult{ReportID: "rep-1", Status: pdp.PlatformSubmitted}}
	srv := actions.NewServer(db, nil, nil, nil, nil, nil, mock)
	resp, err := srv.SubmitInvoiceReport(context.Background(), &invoiceGrpc.SubmitInvoiceReportRequest{
		UserId: userID, Kind: "TRANSACTION", Year: 2099, Month: 4,
	})
	if err != nil || !resp.Success || resp.Code != codes.Success {
		t.Fatalf("submit: err=%v success=%v code=%d", err, resp.Success, resp.Code)
	}
	status, ht, vat, rid := reportRow(t, db, userID, pdp.ReportTransaction, 2099, 4)
	if status != "DEPOSITED" || rid.String != "rep-1" {
		t.Fatalf("status=%q report_id=%q; want DEPOSITED/rep-1", status, rid.String)
	}
	if ht != 150_000 || vat != 30_000 {
		t.Fatalf("totals HT=%d VAT=%d; want 150000/30000 (FR April only)", ht, vat)
	}
	if len(mock.Reports) != 1 || mock.Reports[0].TotalHTCents != 150_000 {
		t.Fatalf("platform Reports=%+v; want one with HT 150000", mock.Reports)
	}
}

// INV-BE-091: a CROSS_BORDER_B2C report aggregates the intra-EU OSS assiette only,
// net of credit notes, and never the domestic sales.
func TestSubmitReport_CrossBorderNetsCreditNotes(t *testing.T) {
	db := sealTestDB(t)
	const userID = "rep-cb"
	apr := time.Date(2099, 4, 10, 9, 0, 0, 0, time.UTC)

	seedReportInvoice(t, db, userID, "inv-de-1", "DE", apr, 300_000, 57_000, true)
	// Domestic sale: excluded from the cross-border report.
	seedReportInvoice(t, db, userID, "inv-fr", "FR", apr, 999_999, 0, false)
	// Credit note neutralising part of the DE sale, same scope/month: deducted.
	seedReportCreditNote(t, db, userID, "cn-de", "inv-de-1", "DE", apr, 100_000, 19_000, true)

	mock := &pdp.MockReporter{SubmitResult: pdp.SubmitReportResult{ReportID: "rep-cb", Status: pdp.PlatformSubmitted}}
	srv := actions.NewServer(db, nil, nil, nil, nil, nil, mock)
	resp, err := srv.SubmitInvoiceReport(context.Background(), &invoiceGrpc.SubmitInvoiceReportRequest{
		UserId: userID, Kind: "CROSS_BORDER_B2C", Year: 2099, Month: 4,
	})
	if err != nil || !resp.Success {
		t.Fatalf("submit: err=%v success=%v code=%d", err, resp.Success, resp.Code)
	}
	_, ht, vat, _ := reportRow(t, db, userID, pdp.ReportCrossBorderB2C, 2099, 4)
	if ht != 200_000 || vat != 38_000 {
		t.Fatalf("totals HT=%d VAT=%d; want 200000/38000 (DE net of credit note)", ht, vat)
	}
}

// INV-BE-092: re-submitting a non-terminal period replaces the aggregate
// (idempotent UPSERT) instead of inserting a duplicate.
func TestSubmitReport_ResubmitUpserts(t *testing.T) {
	db := sealTestDB(t)
	const userID = "rep-upsert"
	apr := time.Date(2099, 4, 10, 9, 0, 0, 0, time.UTC)
	seedReportInvoice(t, db, userID, "inv-fr-1", "FR", apr, 100_000, 20_000, false)

	mock := &pdp.MockReporter{SubmitResult: pdp.SubmitReportResult{ReportID: "rep-1", Status: pdp.PlatformSubmitted}}
	srv := actions.NewServer(db, nil, nil, nil, nil, nil, mock)
	ctx := context.Background()
	req := &invoiceGrpc.SubmitInvoiceReportRequest{UserId: userID, Kind: "TRANSACTION", Year: 2099, Month: 4}
	if r, _ := srv.SubmitInvoiceReport(ctx, req); !r.Success {
		t.Fatalf("first submit failed: code=%d", r.Code)
	}
	// A new sale lands in the same period, then we re-submit.
	seedReportInvoice(t, db, userID, "inv-fr-2", "FR", apr, 25_000, 5_000, false)
	if r, _ := srv.SubmitInvoiceReport(ctx, req); !r.Success {
		t.Fatalf("re-submit failed: code=%d", r.Code)
	}

	var count int
	if err := db.QueryRow(
		`SELECT COUNT(*) FROM invoice_reports WHERE user_id=$1 AND kind='TRANSACTION' AND period_year=2099 AND period_month=4`,
		userID,
	).Scan(&count); err != nil {
		t.Fatalf("count: %v", err)
	}
	if count != 1 {
		t.Fatalf("report rows=%d; want 1 (upsert, not duplicate)", count)
	}
	_, ht, _, _ := reportRow(t, db, userID, pdp.ReportTransaction, 2099, 4)
	if ht != 125_000 {
		t.Fatalf("HT=%d; want 125000 (fresh aggregate after resubmit)", ht)
	}
}

// INV-BE-093: the report poller advances a submitted report's status through the
// strict B3 path, one cran at a time, from the platform-reported status.
func TestReportPoller_AdvancesStatus(t *testing.T) {
	db := sealTestDB(t)
	const userID = "rep-poll"
	apr := time.Date(2099, 4, 10, 9, 0, 0, 0, time.UTC)
	seedReportInvoice(t, db, userID, "inv-fr-1", "FR", apr, 100_000, 20_000, false)

	mock := &pdp.MockReporter{SubmitResult: pdp.SubmitReportResult{ReportID: "rep-1", Status: pdp.PlatformSubmitted}}
	srv := actions.NewServer(db, nil, nil, nil, nil, nil, mock)
	ctx := context.Background()
	if r, _ := srv.SubmitInvoiceReport(ctx, &invoiceGrpc.SubmitInvoiceReportRequest{
		UserId: userID, Kind: "TRANSACTION", Year: 2099, Month: 4,
	}); !r.Success {
		t.Fatalf("submit failed: code=%d", r.Code)
	}

	// Platform now reports APPROVED: the poller must walk DEPOSITED→RECEIVED→APPROVED.
	mock.StatusResult = pdp.PlatformApproved
	srv.PollReportStatuses(ctx)
	if status, _, _, _ := reportRow(t, db, userID, pdp.ReportTransaction, 2099, 4); status != "APPROVED" {
		t.Fatalf("status=%q; want APPROVED after poll", status)
	}
}
