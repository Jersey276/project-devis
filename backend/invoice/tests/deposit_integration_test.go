package tests

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"project-devis-invoice/actions"
	"project-devis-invoice/actions/codes"
	"project-devis-invoice/pdp"
	invoiceGrpc "project-devis-invoice/services/grpc"

	_ "github.com/lib/pq"
)

func pdpSubmissionID(t *testing.T, db *sql.DB, invoiceID string) string {
	t.Helper()
	var sid sql.NullString
	if err := db.QueryRow(
		`SELECT pdp_submission_id FROM invoices WHERE invoice_id=$1`, invoiceID,
	).Scan(&sid); err != nil {
		t.Fatalf("read pdp_submission_id: %v", err)
	}
	return sid.String
}

func TestDeposit_DrivesDEPOSITED(t *testing.T) {
	db := sealTestDB(t)
	const userID = "deposit-ok"
	seedIssuedInvoice(t, db, userID, "inv-dep", "2099-0001",
		time.Date(2099, 4, 1, 9, 0, 0, 0, time.UTC), 1)

	mock := &pdp.MockClient{SubmitResult: pdp.SubmitResult{SubmissionID: "sub-1", Status: pdp.PlatformSubmitted}}
	srv := actions.NewServer(db, nil, nil, nil, mock, nil)
	resp, err := srv.DepositInvoice(context.Background(), &invoiceGrpc.DepositInvoiceRequest{
		InvoiceId: "inv-dep", UserId: userID, Note: "dépôt initial",
	})
	if err != nil {
		t.Fatalf("deposit: %v", err)
	}
	if !resp.Success || resp.Code != codes.Success {
		t.Fatalf("deposit: success=%v code=%d; want success/0", resp.Success, resp.Code)
	}
	if got := lifecycleState(t, db, "inv-dep"); got != "DEPOSITED" {
		t.Fatalf("lifecycle_status=%q; want DEPOSITED", got)
	}
	if n := lifecycleEventCount(t, db, "inv-dep"); n != 1 {
		t.Fatalf("event count=%d; want 1", n)
	}
	if got := pdpSubmissionID(t, db, "inv-dep"); got != "sub-1" {
		t.Fatalf("pdp_submission_id=%q; want sub-1", got)
	}
	if len(mock.Submitted) != 1 || mock.Submitted[0].InvoiceNumber != "2099-0001" {
		t.Fatalf("platform Submit calls=%+v; want one with invoice number 2099-0001", mock.Submitted)
	}
}

func TestDeposit_DraftRefused(t *testing.T) {
	db := sealTestDB(t)
	const userID = "deposit-draft"
	seedDraftInvoice(t, db, userID, "inv-dep-draft")

	mock := &pdp.MockClient{SubmitResult: pdp.SubmitResult{Status: pdp.PlatformSubmitted}}
	srv := actions.NewServer(db, nil, nil, nil, mock, nil)
	resp, err := srv.DepositInvoice(context.Background(), &invoiceGrpc.DepositInvoiceRequest{
		InvoiceId: "inv-dep-draft", UserId: userID,
	})
	if err != nil {
		t.Fatalf("deposit: %v", err)
	}
	if resp.Success || resp.Code != codes.LifecycleRequiresIssued {
		t.Fatalf("success=%v code=%d; want LifecycleRequiresIssued (%d)", resp.Success, resp.Code, codes.LifecycleRequiresIssued)
	}
	// A draft must never reach the platform.
	if len(mock.Submitted) != 0 {
		t.Fatalf("platform was called for a draft: %+v", mock.Submitted)
	}
}

func TestDeposit_DoubleDepositRefused(t *testing.T) {
	db := sealTestDB(t)
	const userID = "deposit-twice"
	seedIssuedInvoice(t, db, userID, "inv-dep-2", "2099-0001",
		time.Date(2099, 4, 1, 9, 0, 0, 0, time.UTC), 1)

	mock := &pdp.MockClient{SubmitResult: pdp.SubmitResult{SubmissionID: "sub-1", Status: pdp.PlatformSubmitted}}
	srv := actions.NewServer(db, nil, nil, nil, mock, nil)
	ctx := context.Background()
	first, err := srv.DepositInvoice(ctx, &invoiceGrpc.DepositInvoiceRequest{InvoiceId: "inv-dep-2", UserId: userID})
	if err != nil || !first.Success {
		t.Fatalf("first deposit: err=%v success=%v code=%d", err, first.Success, first.Code)
	}
	second, err := srv.DepositInvoice(ctx, &invoiceGrpc.DepositInvoiceRequest{InvoiceId: "inv-dep-2", UserId: userID})
	if err != nil {
		t.Fatalf("second deposit: %v", err)
	}
	if second.Success || second.Code != codes.LifecycleTransitionInvalid {
		t.Fatalf("second deposit: success=%v code=%d; want LifecycleTransitionInvalid (%d)", second.Success, second.Code, codes.LifecycleTransitionInvalid)
	}
	if n := lifecycleEventCount(t, db, "inv-dep-2"); n != 1 {
		t.Fatalf("event count=%d; want 1 (second deposit must not append)", n)
	}
}

func TestDeposit_PlatformErrorLeavesStateUntouched(t *testing.T) {
	db := sealTestDB(t)
	const userID = "deposit-err"
	seedIssuedInvoice(t, db, userID, "inv-dep-err", "2099-0001",
		time.Date(2099, 4, 1, 9, 0, 0, 0, time.UTC), 1)

	mock := &pdp.MockClient{SubmitErr: errors.New("platform unreachable")}
	srv := actions.NewServer(db, nil, nil, nil, mock, nil)
	resp, err := srv.DepositInvoice(context.Background(), &invoiceGrpc.DepositInvoiceRequest{
		InvoiceId: "inv-dep-err", UserId: userID,
	})
	if err != nil {
		t.Fatalf("deposit: %v", err)
	}
	if resp.Success || resp.Code != codes.PDPSubmissionFailed {
		t.Fatalf("success=%v code=%d; want PDPSubmissionFailed (%d)", resp.Success, resp.Code, codes.PDPSubmissionFailed)
	}
	if got := lifecycleState(t, db, "inv-dep-err"); got != "NONE" {
		t.Fatalf("lifecycle_status=%q; want NONE (untouched)", got)
	}
	if n := lifecycleEventCount(t, db, "inv-dep-err"); n != 0 {
		t.Fatalf("event count=%d; want 0", n)
	}
}

func TestDeposit_WrongOwnerNotFound(t *testing.T) {
	db := sealTestDB(t)
	seedIssuedInvoice(t, db, "owner-a", "inv-dep-own", "2099-0001",
		time.Date(2099, 4, 1, 9, 0, 0, 0, time.UTC), 1)

	mock := &pdp.MockClient{SubmitResult: pdp.SubmitResult{Status: pdp.PlatformSubmitted}}
	srv := actions.NewServer(db, nil, nil, nil, mock, nil)
	resp, err := srv.DepositInvoice(context.Background(), &invoiceGrpc.DepositInvoiceRequest{
		InvoiceId: "inv-dep-own", UserId: "owner-b",
	})
	if err != nil {
		t.Fatalf("deposit: %v", err)
	}
	if resp.Success || resp.Code != codes.NotFound {
		t.Fatalf("success=%v code=%d; want NotFound (%d)", resp.Success, resp.Code, codes.NotFound)
	}
	if len(mock.Submitted) != 0 {
		t.Fatalf("platform was called for a non-owned invoice: %+v", mock.Submitted)
	}
}
