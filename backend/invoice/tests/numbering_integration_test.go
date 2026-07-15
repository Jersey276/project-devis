package tests

import (
	"context"
	"sync"
	"testing"

	"project-devis-invoice/actions"

	_ "github.com/lib/pq"
)

func TestAllocateInvoiceNumber_Concurrent(t *testing.T) {
	db := sealTestDB(t)

	ctx := context.Background()
	const userID = "concurrent-test-user"
	const year = 2099

	const n = 50
	results := make([]int, n)
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func(idx int) {
			defer wg.Done()
			tx, err := db.BeginTx(ctx, nil)
			if err != nil {
				t.Errorf("begin: %v", err)
				return
			}
			_, seq, err := actions.AllocateInvoiceNumberForTest(ctx, tx, userID, year)
			if err != nil {
				_ = tx.Rollback()
				t.Errorf("allocate: %v", err)
				return
			}
			if err := tx.Commit(); err != nil {
				t.Errorf("commit: %v", err)
				return
			}
			results[idx] = seq
		}(i)
	}
	wg.Wait()

	seen := make(map[int]bool, n)
	for _, seq := range results {
		if seq < 1 || seq > n {
			t.Fatalf("sequence out of range: %d", seq)
		}
		if seen[seq] {
			t.Fatalf("duplicate sequence value: %d", seq)
		}
		seen[seq] = true
	}

	for i := 1; i <= n; i++ {
		if !seen[i] {
			t.Fatalf("gap in sequence: %d missing", i)
		}
	}
}
