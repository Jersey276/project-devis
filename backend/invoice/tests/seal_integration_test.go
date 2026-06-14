package tests

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"

	"project-devis-invoice/actions"
	invoiceGrpc "project-devis-invoice/services/grpc"

	_ "github.com/lib/pq"
)

// These tests exercise the real chaining, triggers, backfill and verification
// against a disposable Postgres. Skipped unless INVOICE_TEST_DATABASE_URL is set
// to a database owned by role devis-invoice (needs schema-create privilege).
//
//	INVOICE_TEST_DATABASE_URL="postgres://devis-invoice:pass@localhost:5432/invoice?sslmode=disable" go test ./tests/ -run Seal
//
// Each test runs in its own disposable schema: the helper creates a fresh schema,
// applies every migration into it, and drops it (CASCADE) on cleanup. This needs
// no superuser, leaves the database untouched, and is fully re-runnable — DROP
// SCHEMA CASCADE removes the tables outright, so the append-only triggers (which
// only fire on row UPDATE/DELETE) never get in the way.

// sealTestDB returns a *sql.DB whose every pooled connection is pinned to a fresh,
// migrated, disposable schema. The schema is dropped on test cleanup.
func sealTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dsn := os.Getenv("INVOICE_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("set INVOICE_TEST_DATABASE_URL to run the seal integration tests")
	}

	// Unique schema name for this test run (lowercase, valid identifier).
	schema := fmt.Sprintf("sealtest_%d", time.Now().UnixNano())

	// 1. Create the schema on a short-lived admin connection.
	admin, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Fatalf("open admin db: %v", err)
	}
	if err := admin.Ping(); err != nil {
		admin.Close()
		t.Fatalf("ping db: %v", err)
	}
	if _, err := admin.Exec(fmt.Sprintf(`CREATE SCHEMA %q`, schema)); err != nil {
		admin.Close()
		t.Fatalf("create schema %s: %v", schema, err)
	}
	admin.Close()

	// 2. Open the test pool with search_path pinned to the schema, so every
	//    connection — and thus the prod code under test — resolves tables there.
	db, err := sql.Open("postgres", withSearchPath(dsn, schema))
	if err != nil {
		dropSchema(t, dsn, schema)
		t.Fatalf("open test db: %v", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		dropSchema(t, dsn, schema)
		t.Fatalf("ping test db: %v", err)
	}

	// 3. Apply every migration into the schema.
	if err := applyMigrations(db); err != nil {
		db.Close()
		dropSchema(t, dsn, schema)
		t.Fatalf("apply migrations into %s: %v", schema, err)
	}

	t.Cleanup(func() {
		db.Close()
		dropSchema(t, dsn, schema)
	})
	return db
}

// withSearchPath appends a libpq `options=-c search_path=<schema>` parameter to
// the DSN so the pinned schema applies to every connection in the pool.
func withSearchPath(dsn, schema string) string {
	opt := "options=-c search_path=" + schema
	if strings.Contains(dsn, "?") {
		return dsn + "&" + opt
	}
	return dsn + "?" + opt
}

// dropSchema tears down the disposable schema on its own admin connection.
func dropSchema(t *testing.T, dsn, schema string) {
	t.Helper()
	admin, err := sql.Open("postgres", dsn)
	if err != nil {
		t.Logf("drop schema %s: open: %v", schema, err)
		return
	}
	defer admin.Close()
	if _, err := admin.Exec(fmt.Sprintf(`DROP SCHEMA IF EXISTS %q CASCADE`, schema)); err != nil {
		t.Logf("drop schema %s: %v", schema, err)
	}
}

// applyMigrations runs every *.up.sql migration, in filename order, on db. db's
// search_path already points at the disposable schema, so all objects land there.
func applyMigrations(db *sql.DB) error {
	dir := filepath.Join("..", "migrations")
	entries, err := os.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("read migrations dir: %w", err)
	}
	var ups []string
	for _, e := range entries {
		if strings.HasSuffix(e.Name(), ".up.sql") {
			ups = append(ups, e.Name())
		}
	}
	sort.Strings(ups)
	for _, name := range ups {
		sqlBytes, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return fmt.Errorf("read %s: %w", name, err)
		}
		if _, err := db.Exec(string(sqlBytes)); err != nil {
			return fmt.Errorf("exec %s: %w", name, err)
		}
	}
	return nil
}

// seedIssuedInvoice inserts an ISSUED invoice + one line snapshot directly,
// bypassing the gRPC issue path. Returns the invoice id.
func seedIssuedInvoice(t *testing.T, db *sql.DB, userID, invoiceID, number string, issuedAt time.Time, seq int) {
	t.Helper()
	_, err := db.Exec(
		`INSERT INTO invoices (invoice_id, user_id, quote_id, status, invoice_number, number_year, number_seq,
		                       issued_at, total_ht_cents, total_vat_cents, total_ttc_cents, vat_exempt)
		 VALUES ($1,$2,'quote-x','ISSUED',$3,$4,$5,$6,10000,2000,12000,false)`,
		invoiceID, userID, number, issuedAt.Year(), seq, issuedAt)
	if err != nil {
		t.Fatalf("seed invoice: %v", err)
	}
	_, err = db.Exec(
		`INSERT INTO invoice_line_snapshots (invoice_id, position, quote_line_id, name, unit, quantity,
		                                     unit_price_cents, line_ht_cents, tax_id, tax_rate, tax_label)
		 VALUES ($1,0,'line-1','Prestation','u','1',10000,10000,0,'20','TVA 20%')`,
		invoiceID)
	if err != nil {
		t.Fatalf("seed line: %v", err)
	}
}

func TestSeal_BackfillAndVerify(t *testing.T) {
	db := sealTestDB(t)
	const userID = "seal-test-backfill"

	base := time.Date(2099, 1, 10, 9, 0, 0, 0, time.UTC)
	seedIssuedInvoice(t, db, userID, "inv-a", "2099-0001", base, 1)
	seedIssuedInvoice(t, db, userID, "inv-b", "2099-0002", base.Add(time.Hour), 2)

	if err := actions.BackfillSeals(context.Background(), db); err != nil {
		t.Fatalf("backfill: %v", err)
	}

	// Two contiguous seals, indexes 0 and 1.
	var count int
	if err := db.QueryRow(`SELECT count(*) FROM document_seals WHERE user_id=$1`, userID).Scan(&count); err != nil {
		t.Fatalf("count seals: %v", err)
	}
	if count != 2 {
		t.Fatalf("seals = %d; want 2", count)
	}

	srv := actions.NewServer(db, nil, nil, nil)
	resp, err := srv.VerifyChain(context.Background(), &invoiceGrpc.VerifyChainRequest{UserId: userID})
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !resp.Ok || resp.Checked != 2 {
		t.Fatalf("verify ok=%v checked=%d reason=%q; want ok/2", resp.Ok, resp.Checked, resp.Reason)
	}

	// Idempotent re-run.
	if err := actions.BackfillSeals(context.Background(), db); err != nil {
		t.Fatalf("backfill re-run: %v", err)
	}
	if err := db.QueryRow(`SELECT count(*) FROM document_seals WHERE user_id=$1`, userID).Scan(&count); err != nil {
		t.Fatalf("count seals after re-run: %v", err)
	}
	if count != 2 {
		t.Fatalf("seals after re-run = %d; want 2 (idempotent)", count)
	}
}

func TestSeal_TriggerBlocksTamper(t *testing.T) {
	db := sealTestDB(t)
	const userID = "seal-test-trigger"

	seedIssuedInvoice(t, db, userID, "inv-t", "2099-0001", time.Date(2099, 2, 1, 9, 0, 0, 0, time.UTC), 1)
	if err := actions.BackfillSeals(context.Background(), db); err != nil {
		t.Fatalf("backfill: %v", err)
	}

	// Direct tamper on a sealed invoice must be blocked by the trigger.
	if _, err := db.Exec(`UPDATE invoices SET total_ttc_cents = total_ttc_cents + 1 WHERE invoice_id='inv-t'`); err == nil {
		t.Fatal("UPDATE on sealed invoice succeeded; trigger did not block it")
	}
	// Tampering a line snapshot must be blocked.
	if _, err := db.Exec(`UPDATE invoice_line_snapshots SET line_ht_cents = 1 WHERE invoice_id='inv-t'`); err == nil {
		t.Fatal("UPDATE on line snapshot succeeded; trigger did not block it")
	}
	// document_seals is append-only.
	if _, err := db.Exec(`UPDATE document_seals SET chain_hash='x' WHERE doc_id='inv-t'`); err == nil {
		t.Fatal("UPDATE on document_seals succeeded; trigger did not block it")
	}
	if _, err := db.Exec(`DELETE FROM invoices WHERE invoice_id='inv-t'`); err == nil {
		t.Fatal("DELETE of sealed invoice succeeded; trigger did not block it")
	}
}

func TestSeal_MarkInvoicePaidStillAllowed(t *testing.T) {
	db := sealTestDB(t)
	const userID = "seal-test-paid"

	seedIssuedInvoice(t, db, userID, "inv-p", "2099-0001", time.Date(2099, 3, 1, 9, 0, 0, 0, time.UTC), 1)
	if err := actions.BackfillSeals(context.Background(), db); err != nil {
		t.Fatalf("backfill: %v", err)
	}

	// The ISSUED -> PAID transition is whitelisted by the trigger.
	srv := actions.NewServer(db, nil, nil, nil)
	resp, err := srv.MarkInvoicePaid(context.Background(), &invoiceGrpc.MarkInvoicePaidRequest{
		InvoiceId: "inv-p", UserId: userID,
	})
	if err != nil {
		t.Fatalf("mark paid error: %v", err)
	}
	if !resp.Success {
		t.Fatalf("mark paid failed with code %d; the trigger should allow ISSUED->PAID", resp.Code)
	}

	var status string
	if err := db.QueryRow(`SELECT status FROM invoices WHERE invoice_id='inv-p'`).Scan(&status); err != nil {
		t.Fatalf("read status: %v", err)
	}
	if status != "PAID" {
		t.Fatalf("status = %q; want PAID", status)
	}
}
