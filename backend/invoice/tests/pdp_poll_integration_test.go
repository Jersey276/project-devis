package tests

import (
	"context"
	"testing"
	"time"

	"project-devis-invoice/actions"
	"project-devis-invoice/pdp"
	invoiceGrpc "project-devis-invoice/services/grpc"

	_ "github.com/lib/pq"
)

func TestPoll_AdvancesThroughPlatformStatus(t *testing.T) {
	db := sealTestDB(t)
	const userID = "poll-advance"
	seedIssuedInvoice(t, db, userID, "inv-poll", "2099-0001",
		time.Date(2099, 5, 1, 9, 0, 0, 0, time.UTC), 1)

	mock := &pdp.MockClient{SubmitResult: pdp.SubmitResult{SubmissionID: "sub-poll", Status: pdp.PlatformSubmitted}}
	srv := actions.NewServer(db, nil, nil, nil, mock, nil)
	ctx := context.Background()

	if resp, err := srv.DepositInvoice(ctx, &invoiceGrpc.DepositInvoiceRequest{InvoiceId: "inv-poll", UserId: userID}); err != nil || !resp.Success {
		t.Fatalf("deposit: err=%v success=%v code=%d", err, resp.GetSuccess(), resp.GetCode())
	}
	if got := lifecycleState(t, db, "inv-poll"); got != "DEPOSITED" {
		t.Fatalf("after deposit lifecycle=%q; want DEPOSITED", got)
	}

	// Platform reports APPROVED: poller must walk DEPOSITED→RECEIVED→APPROVED.
	mock.StatusResult = pdp.PlatformApproved
	srv.PollPDPStatuses(ctx)

	if got := lifecycleState(t, db, "inv-poll"); got != "APPROVED" {
		t.Fatalf("after poll lifecycle=%q; want APPROVED", got)
	}
	// One event for the deposit + two crans walked = three rows.
	if n := lifecycleEventCount(t, db, "inv-poll"); n != 3 {
		t.Fatalf("event count=%d; want 3 (DEPOSITED, RECEIVED, APPROVED)", n)
	}

	// A second sweep with no platform change is a no-op (idempotent).
	srv.PollPDPStatuses(ctx)
	if n := lifecycleEventCount(t, db, "inv-poll"); n != 3 {
		t.Fatalf("event count after re-poll=%d; want 3 (no duplicate)", n)
	}

	// Platform then reports COLLECTED: one further cran.
	mock.StatusResult = pdp.PlatformCollected
	srv.PollPDPStatuses(ctx)
	if got := lifecycleState(t, db, "inv-poll"); got != "COLLECTED" {
		t.Fatalf("after collect poll lifecycle=%q; want COLLECTED", got)
	}
	if n := lifecycleEventCount(t, db, "inv-poll"); n != 4 {
		t.Fatalf("event count=%d; want 4", n)
	}
}

func TestPoll_RejectedFromDeposited(t *testing.T) {
	db := sealTestDB(t)
	const userID = "poll-reject"
	seedIssuedInvoice(t, db, userID, "inv-rej", "2099-0001",
		time.Date(2099, 5, 1, 9, 0, 0, 0, time.UTC), 1)

	mock := &pdp.MockClient{SubmitResult: pdp.SubmitResult{SubmissionID: "sub-rej", Status: pdp.PlatformSubmitted}}
	srv := actions.NewServer(db, nil, nil, nil, mock, nil)
	ctx := context.Background()
	if resp, err := srv.DepositInvoice(ctx, &invoiceGrpc.DepositInvoiceRequest{InvoiceId: "inv-rej", UserId: userID}); err != nil || !resp.Success {
		t.Fatalf("deposit: err=%v success=%v", err, resp.GetSuccess())
	}

	mock.StatusResult = pdp.PlatformRejected
	srv.PollPDPStatuses(ctx)

	if got := lifecycleState(t, db, "inv-rej"); got != "REJECTED" {
		t.Fatalf("after poll lifecycle=%q; want REJECTED", got)
	}
}

func TestPoll_UnknownLeavesUntouched(t *testing.T) {
	db := sealTestDB(t)
	const userID = "poll-unknown"
	seedIssuedInvoice(t, db, userID, "inv-unk", "2099-0001",
		time.Date(2099, 5, 1, 9, 0, 0, 0, time.UTC), 1)

	mock := &pdp.MockClient{SubmitResult: pdp.SubmitResult{SubmissionID: "sub-unk", Status: pdp.PlatformSubmitted}}
	srv := actions.NewServer(db, nil, nil, nil, mock, nil)
	ctx := context.Background()
	if resp, err := srv.DepositInvoice(ctx, &invoiceGrpc.DepositInvoiceRequest{InvoiceId: "inv-unk", UserId: userID}); err != nil || !resp.Success {
		t.Fatalf("deposit: err=%v success=%v", err, resp.GetSuccess())
	}

	// No-op-style UNKNOWN: poller must not move the lifecycle (production default).
	mock.StatusResult = pdp.PlatformUnknown
	srv.PollPDPStatuses(ctx)

	if got := lifecycleState(t, db, "inv-unk"); got != "DEPOSITED" {
		t.Fatalf("after poll lifecycle=%q; want DEPOSITED (untouched)", got)
	}
	if n := lifecycleEventCount(t, db, "inv-unk"); n != 1 {
		t.Fatalf("event count=%d; want 1 (deposit only)", n)
	}
}

func TestPoll_SkipsInvoicesWithoutSubmissionID(t *testing.T) {
	db := sealTestDB(t)
	const userID = "poll-no-sub"
	seedIssuedInvoice(t, db, userID, "inv-nosub", "2099-0001",
		time.Date(2099, 5, 1, 9, 0, 0, 0, time.UTC), 1)

	// Deposited via the no-op adapter: DEPOSITED but pdp_submission_id stays NULL.
	srv := actions.NewServer(db, nil, nil, nil, pdp.NoopClient{}, nil)
	ctx := context.Background()
	if resp, err := srv.DepositInvoice(ctx, &invoiceGrpc.DepositInvoiceRequest{InvoiceId: "inv-nosub", UserId: userID}); err != nil || !resp.Success {
		t.Fatalf("deposit: err=%v success=%v", err, resp.GetSuccess())
	}

	// A mock that would advance — but the row has no submission id, so it must be
	// excluded from the sweep entirely.
	mock := &pdp.MockClient{StatusResult: pdp.PlatformApproved}
	pollSrv := actions.NewServer(db, nil, nil, nil, mock, nil)
	pollSrv.PollPDPStatuses(ctx)

	if got := lifecycleState(t, db, "inv-nosub"); got != "DEPOSITED" {
		t.Fatalf("after poll lifecycle=%q; want DEPOSITED (no submission id => skipped)", got)
	}
	if len(mock.Submitted) != 0 {
		t.Fatalf("unexpected Submit calls during poll: %+v", mock.Submitted)
	}
}
