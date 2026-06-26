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

func lifecycleState(t *testing.T, db *sql.DB, invoiceID string) string {
	t.Helper()
	var lifecycle string
	if err := db.QueryRow(
		`SELECT lifecycle_status FROM invoices WHERE invoice_id=$1`, invoiceID,
	).Scan(&lifecycle); err != nil {
		t.Fatalf("read lifecycle_status: %v", err)
	}
	return lifecycle
}

func lifecycleEventCount(t *testing.T, db *sql.DB, invoiceID string) int {
	t.Helper()
	var n int
	if err := db.QueryRow(
		`SELECT COUNT(*) FROM invoice_lifecycle_events WHERE invoice_id=$1`, invoiceID,
	).Scan(&n); err != nil {
		t.Fatalf("count events: %v", err)
	}
	return n
}

func TestSetLifecycle_TransitionAppendsEvent(t *testing.T) {
	db := sealTestDB(t)
	const userID = "lifecycle-ok"
	seedIssuedInvoice(t, db, userID, "inv-lc", "2099-0001",
		time.Date(2099, 4, 1, 9, 0, 0, 0, time.UTC), 1)

	srv := actions.NewServer(db, nil, nil, nil, nil, nil, nil)
	resp, err := srv.SetInvoiceLifecycleStatus(context.Background(), &invoiceGrpc.SetInvoiceLifecycleStatusRequest{
		InvoiceId: "inv-lc", UserId: userID, Status: "DEPOSITED", Note: "déposée sur le PDP",
	})
	if err != nil {
		t.Fatalf("set lifecycle: %v", err)
	}
	if !resp.Success || resp.Code != codes.Success {
		t.Fatalf("set lifecycle: success=%v code=%d; want success/0", resp.Success, resp.Code)
	}
	if got := lifecycleState(t, db, "inv-lc"); got != "DEPOSITED" {
		t.Fatalf("lifecycle_status=%q; want DEPOSITED", got)
	}
	if n := lifecycleEventCount(t, db, "inv-lc"); n != 1 {
		t.Fatalf("event count=%d; want 1", n)
	}
}

func TestSetLifecycle_InvalidTransitionLeavesStateUntouched(t *testing.T) {
	db := sealTestDB(t)
	const userID = "lifecycle-bad"
	seedIssuedInvoice(t, db, userID, "inv-bad", "2099-0001",
		time.Date(2099, 4, 1, 9, 0, 0, 0, time.UTC), 1)

	srv := actions.NewServer(db, nil, nil, nil, nil, nil, nil)
	// NONE → APPROVED skips two hops; must be refused.
	resp, err := srv.SetInvoiceLifecycleStatus(context.Background(), &invoiceGrpc.SetInvoiceLifecycleStatusRequest{
		InvoiceId: "inv-bad", UserId: userID, Status: "APPROVED",
	})
	if err != nil {
		t.Fatalf("set lifecycle: %v", err)
	}
	if resp.Success || resp.Code != codes.LifecycleTransitionInvalid {
		t.Fatalf("success=%v code=%d; want LifecycleTransitionInvalid (%d)", resp.Success, resp.Code, codes.LifecycleTransitionInvalid)
	}
	if got := lifecycleState(t, db, "inv-bad"); got != "NONE" {
		t.Fatalf("lifecycle_status=%q; want NONE (untouched)", got)
	}
	if n := lifecycleEventCount(t, db, "inv-bad"); n != 0 {
		t.Fatalf("event count=%d; want 0", n)
	}
}

func TestSetLifecycle_DraftRefused(t *testing.T) {
	db := sealTestDB(t)
	const userID = "lifecycle-draft"
	seedDraftInvoice(t, db, userID, "inv-draft-lc")

	srv := actions.NewServer(db, nil, nil, nil, nil, nil, nil)
	resp, err := srv.SetInvoiceLifecycleStatus(context.Background(), &invoiceGrpc.SetInvoiceLifecycleStatusRequest{
		InvoiceId: "inv-draft-lc", UserId: userID, Status: "DEPOSITED",
	})
	if err != nil {
		t.Fatalf("set lifecycle: %v", err)
	}
	if resp.Success || resp.Code != codes.LifecycleRequiresIssued {
		t.Fatalf("success=%v code=%d; want LifecycleRequiresIssued (%d)", resp.Success, resp.Code, codes.LifecycleRequiresIssued)
	}
}

func TestSetLifecycle_WrongOwnerNotFound(t *testing.T) {
	db := sealTestDB(t)
	seedIssuedInvoice(t, db, "owner-a", "inv-own", "2099-0001",
		time.Date(2099, 4, 1, 9, 0, 0, 0, time.UTC), 1)

	srv := actions.NewServer(db, nil, nil, nil, nil, nil, nil)
	resp, err := srv.SetInvoiceLifecycleStatus(context.Background(), &invoiceGrpc.SetInvoiceLifecycleStatusRequest{
		InvoiceId: "inv-own", UserId: "owner-b", Status: "DEPOSITED",
	})
	if err != nil {
		t.Fatalf("set lifecycle: %v", err)
	}
	if resp.Success || resp.Code != codes.NotFound {
		t.Fatalf("success=%v code=%d; want NotFound (%d)", resp.Success, resp.Code, codes.NotFound)
	}
}

func TestListLifecycleEvents_OrderedHistory(t *testing.T) {
	db := sealTestDB(t)
	const userID = "lifecycle-hist"
	seedIssuedInvoice(t, db, userID, "inv-hist", "2099-0001",
		time.Date(2099, 4, 1, 9, 0, 0, 0, time.UTC), 1)

	srv := actions.NewServer(db, nil, nil, nil, nil, nil, nil)
	ctx := context.Background()
	for _, s := range []string{"DEPOSITED", "RECEIVED", "APPROVED"} {
		resp, err := srv.SetInvoiceLifecycleStatus(ctx, &invoiceGrpc.SetInvoiceLifecycleStatusRequest{
			InvoiceId: "inv-hist", UserId: userID, Status: s,
		})
		if err != nil || !resp.Success {
			t.Fatalf("transition %s: err=%v success=%v code=%d", s, err, resp.Success, resp.Code)
		}
	}

	list, err := srv.ListInvoiceLifecycleEvents(ctx, &invoiceGrpc.ListInvoiceLifecycleEventsRequest{
		InvoiceId: "inv-hist", UserId: userID,
	})
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if !list.Success || len(list.Events) != 3 {
		t.Fatalf("events=%d success=%v; want 3 events", len(list.Events), list.Success)
	}
	want := []string{"DEPOSITED", "RECEIVED", "APPROVED"}
	for i, e := range list.Events {
		if e.Status != want[i] {
			t.Fatalf("event[%d]=%q; want %q", i, e.Status, want[i])
		}
	}
}
