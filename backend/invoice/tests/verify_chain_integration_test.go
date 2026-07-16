package tests

import (
	"context"
	"testing"
	"time"

	"project-devis-invoice/actions"
	invoiceGrpc "project-devis-invoice/services/grpc"
)

// TestVerifyChain_ContentTamperedWithoutTouchingSeal covers the branch of
// VerifyChain that the trigger-based tamper test (TestSeal_TriggerBlocksTamper)
// cannot reach: the chain_hash/prev_hash bookkeeping is internally consistent,
// but the content it commits to doesn't match what's stored. This can't happen
// via UPDATE (blocked by the immutability triggers) but can happen if
// sealDocument itself is ever called with a wrong content_hash (e.g. a bug in
// the caller), which is exactly the class of defect VerifyChain exists to catch.
func TestVerifyChain_ContentTamperedWithoutTouchingSeal(t *testing.T) {
	db := sealTestDB(t)
	const userID = "verify-tamper-test"

	seedIssuedInvoice(t, db, userID, "inv-tamper", "2099-0001", time.Date(2099, 4, 1, 9, 0, 0, 0, time.UTC), 1)

	// Seal it with a content_hash that does NOT match the actual invoice row
	// (as if computeContentHash had been given stale totals at issue time).
	wrongHash := "0000000000000000000000000000000000000000000000000000000000dead"
	chainHash := actions.ComputeChainHashForTest(genesisHashForTest, wrongHash, 0)
	if _, err := db.Exec(
		`INSERT INTO document_seals (user_id, doc_type, doc_id, chain_index, content_hash, prev_hash, chain_hash)
		 VALUES ($1, 'INVOICE', $2, 0, $3, $4, $5)`,
		userID, "inv-tamper", wrongHash, genesisHashForTest, chainHash,
	); err != nil {
		t.Fatalf("seed tampered seal: %v", err)
	}

	srv := actions.NewServer(db, nil, nil, nil, nil, nil, nil)
	resp, err := srv.VerifyChain(context.Background(), &invoiceGrpc.VerifyChainRequest{UserId: userID})
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !resp.Success {
		t.Fatalf("verify call itself should succeed (Success=true), got code=%d", resp.Code)
	}
	if resp.Ok {
		t.Fatal("expected Ok=false: the recomputed content hash should not match the stored one")
	}
	if resp.BrokenDocId != "inv-tamper" || resp.BrokenDocType != "INVOICE" {
		t.Errorf("broken doc = %s/%s; want INVOICE/inv-tamper", resp.BrokenDocType, resp.BrokenDocId)
	}
}

const genesisHashForTest = "0000000000000000000000000000000000000000000000000000000000000000"
