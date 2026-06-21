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

func TestOSSCumulative_SumsAssietteForYear(t *testing.T) {
	db := sealTestDB(t)
	const userID = "oss-cumul"
	srv := actions.NewServer(db, nil, nil, nil)

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
	srv := actions.NewServer(db, nil, nil, nil)

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
	srv := actions.NewServer(db, nil, nil, nil)
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
	if actions.OSSAppliesForTest(false, got, "individual", country("DE", true)) {
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
	if !actions.OSSAppliesForTest(false, got, "individual", country("DE", true)) {
		t.Error("OSS should apply once the threshold is reached")
	}
}
