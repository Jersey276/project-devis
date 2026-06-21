package tests

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"project-devis-invoice/actions"

	_ "github.com/lib/pq"
)

func seedInvoiceHT(t *testing.T, db *sql.DB, userID, invoiceID, status string, htCents int64, issuedAt time.Time, counts bool) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO invoices (invoice_id, user_id, quote_id, status, issued_at,
		                       total_ht_cents, total_vat_cents, total_ttc_cents, vat_exempt)
		 VALUES ($1,$2,'quote-x',$3,$4,$5,0,$5,false)`,
		invoiceID, userID, status, issuedAt, htCents)
	if err != nil {
		t.Fatalf("seed invoice %s: %v", invoiceID, err)
	}
	seedPartySnapshot(t, db, invoiceID, "", "", "individual", 276, counts, "FR", "DE", counts)
}

// seedCreditNoteHT inserts a credit note (with its frozen party snapshot) that
// neutralises htCents of an existing in-scope invoice. counts toggles the
// frozen assiette flag inherited from the origin invoice.
func seedCreditNoteHT(t *testing.T, db *sql.DB, userID, creditNoteID, invoiceID string, htCents int64, issuedAt time.Time, counts bool) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO credit_notes (credit_note_id, user_id, invoice_id, credit_note_number,
		                           number_year, number_seq, issued_at,
		                           total_ht_cents, total_vat_cents, total_ttc_cents, vat_exempt)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,0,$8,false)`,
		creditNoteID, userID, invoiceID, "AV-"+creditNoteID, issuedAt.Year(), 1, issuedAt, htCents)
	if err != nil {
		t.Fatalf("seed credit note %s: %v", creditNoteID, err)
	}
	_, err = db.Exec(
		`INSERT INTO credit_note_party_snapshots (credit_note_id, client_type, client_country_id,
		                                          oss_applied, issuer_country_code, client_country_code,
		                                          counts_toward_oss_threshold)
		 VALUES ($1,'individual',276,$2,'FR','DE',$2)`,
		creditNoteID, counts)
	if err != nil {
		t.Fatalf("seed credit note party snapshot %s: %v", creditNoteID, err)
	}
}

func TestOSSCumulative_DeductsCreditNotes(t *testing.T) {
	db := sealTestDB(t)
	const userID = "oss-cn-deduct"
	srv := actions.NewServer(db, nil, nil, nil, nil)

	at := time.Date(2099, 6, 15, 12, 0, 0, 0, time.UTC)
	in2099 := func(month, day int) time.Time { return time.Date(2099, time.Month(month), day, 12, 0, 0, 0, time.UTC) }

	seedInvoiceHT(t, db, userID, "inv-1", "ISSUED", 800_000, in2099(1, 10), true)
	seedInvoiceHT(t, db, userID, "inv-2", "PAID", 500_000, in2099(2, 5), true)

	// In-scope credit note: deducted.
	seedCreditNoteHT(t, db, userID, "cn-1", "inv-1", 300_000, in2099(3, 1), true)
	// Out-of-scope credit note (flag false): ignored.
	seedCreditNoteHT(t, db, userID, "cn-2", "inv-2", 100_000, in2099(3, 2), false)

	got, err := srv.OSSCumulativeHTForYearForTest(context.Background(), userID, "none", at)
	if err != nil {
		t.Fatalf("cumulative: %v", err)
	}
	const want = 800_000 + 500_000 - 300_000
	if got != want {
		t.Errorf("cumulative = %d; want %d (credit notes should be deducted)", got, want)
	}
}

func TestOSSCumulative_SumsAssietteForYear(t *testing.T) {
	db := sealTestDB(t)
	const userID = "oss-cumul"
	srv := actions.NewServer(db, nil, nil, nil, nil)

	at := time.Date(2099, 6, 15, 12, 0, 0, 0, time.UTC)
	in2099 := func(month, day int) time.Time { return time.Date(2099, time.Month(month), day, 12, 0, 0, 0, time.UTC) }

	seedInvoiceHT(t, db, userID, "inv-1", "ISSUED", 400_000, in2099(1, 10), true)
	seedInvoiceHT(t, db, userID, "inv-2", "PAID", 300_000, in2099(3, 5), true)

	seedInvoiceHT(t, db, userID, "inv-3", "ISSUED", 900_000, in2099(2, 1), false)

	seedInvoiceHT(t, db, userID, "inv-4", "CANCELLED", 900_000, in2099(2, 2), true)

	seedInvoiceHT(t, db, userID, "inv-5", "ISSUED", 900_000, time.Date(2098, 6, 30, 12, 0, 0, 0, time.UTC), true)

	seedInvoiceHT(t, db, userID, "inv-draft", "ISSUED", 500_000, in2099(4, 1), true)

	seedInvoiceHT(t, db, "other-user", "inv-x", "ISSUED", 999_999, in2099(5, 1), true)

	got, err := srv.OSSCumulativeHTForYearForTest(context.Background(), userID, "inv-draft", at)
	if err != nil {
		t.Fatalf("cumulative: %v", err)
	}
	const want = 400_000 + 300_000
	if got != want {
		t.Errorf("cumulative = %d; want %d", got, want)
	}
}

func TestOSSCumulative_Empty(t *testing.T) {
	db := sealTestDB(t)
	srv := actions.NewServer(db, nil, nil, nil, nil)

	got, err := srv.OSSCumulativeHTForYearForTest(context.Background(), "nobody", "none",
		time.Date(2099, 6, 15, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("cumulative: %v", err)
	}
	if got != 0 {
		t.Errorf("cumulative = %d; want 0", got)
	}
}

func TestOSSCumulative_ThresholdBoundary(t *testing.T) {
	db := sealTestDB(t)
	const userID = "oss-boundary"
	srv := actions.NewServer(db, nil, nil, nil, nil)
	at := time.Date(2099, 6, 15, 12, 0, 0, 0, time.UTC)

	seedInvoiceHT(t, db, userID, "inv-a", "ISSUED",
		actions.OSSThresholdCentsForTest-1, time.Date(2099, 1, 5, 12, 0, 0, 0, time.UTC), true)

	got, err := srv.OSSCumulativeHTForYearForTest(context.Background(), userID, "none", at)
	if err != nil {
		t.Fatalf("cumulative: %v", err)
	}
	if got != actions.OSSThresholdCentsForTest-1 {
		t.Fatalf("cumulative = %d; want %d", got, actions.OSSThresholdCentsForTest-1)
	}
	if actions.OSSAppliesForTest(false, got, false, "individual", country("DE", true)) {
		t.Error("OSS should not apply just below the threshold without opt-in")
	}

	seedInvoiceHT(t, db, userID, "inv-b", "ISSUED", 100,
		time.Date(2099, 1, 6, 12, 0, 0, 0, time.UTC), true)
	got, err = srv.OSSCumulativeHTForYearForTest(context.Background(), userID, "none", at)
	if err != nil {
		t.Fatalf("cumulative: %v", err)
	}
	if got < actions.OSSThresholdCentsForTest {
		t.Fatalf("cumulative = %d; want >= %d", got, actions.OSSThresholdCentsForTest)
	}
	if !actions.OSSAppliesForTest(false, got, false, "individual", country("DE", true)) {
		t.Error("OSS should apply once the threshold is reached")
	}
}

// TestOSSPriorYear_OverThreshold proves the N-1 rule (art. 259 D CGI): a year
// whose prior civil year crossed the threshold triggers destination VAT from the
// first euro, even with zero current-year turnover. It also checks the two legs
// (current vs prior) are disjoint and that prior-year credit notes are deducted.
func TestOSSPriorYear_OverThreshold(t *testing.T) {
	db := sealTestDB(t)
	const userID = "oss-prior-year"
	srv := actions.NewServer(db, nil, nil, nil, nil)
	ctx := context.Background()
	at := time.Date(2099, 2, 1, 12, 0, 0, 0, time.UTC) // current year N = 2099

	// N-1 (2098) sales crossing the threshold, no N (2099) sale at all.
	seedInvoiceHT(t, db, userID, "inv-prior", "ISSUED",
		actions.OSSThresholdCentsForTest, time.Date(2098, 6, 30, 12, 0, 0, 0, time.UTC), true)

	over, priorCumul, err := srv.OSSPriorYearOverThresholdForTest(ctx, userID, at)
	if err != nil {
		t.Fatalf("prior year: %v", err)
	}
	if !over || priorCumul != actions.OSSThresholdCentsForTest {
		t.Fatalf("prior year over=%v cumul=%d; want true / %d", over, priorCumul, actions.OSSThresholdCentsForTest)
	}
	// Current-year leg is disjoint: zero N sales.
	curr, err := srv.OSSCumulativeHTForYearForTest(ctx, userID, "none", at)
	if err != nil {
		t.Fatalf("current year: %v", err)
	}
	if curr != 0 {
		t.Fatalf("current-year cumulative = %d; want 0 (legs must be disjoint)", curr)
	}
	// Decision: OSS applies despite zero current cumulative.
	if !actions.OSSAppliesForTest(false, curr, over, "individual", country("DE", true)) {
		t.Error("OSS should apply in year N when N-1 crossed the threshold (first-euro rule)")
	}

	// A prior-year credit note that drops N-1 net below the threshold lifts the rule.
	seedCreditNoteHT(t, db, userID, "cn-prior", "inv-prior", 1, time.Date(2098, 7, 1, 12, 0, 0, 0, time.UTC), true)
	over, priorCumul, err = srv.OSSPriorYearOverThresholdForTest(ctx, userID, at)
	if err != nil {
		t.Fatalf("prior year after credit note: %v", err)
	}
	if over || priorCumul != actions.OSSThresholdCentsForTest-1 {
		t.Fatalf("after CN: over=%v cumul=%d; want false / %d", over, priorCumul, actions.OSSThresholdCentsForTest-1)
	}
}
