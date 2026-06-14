package actions

import (
	"context"
	"database/sql"
)

// AllocateCreditNoteNumberForTest exposes allocateCreditNoteNumber to the tests
// package. Thin adapter, no logic of its own.
func AllocateCreditNoteNumberForTest(ctx context.Context, tx *sql.Tx, userID string, year int) (string, int, error) {
	return allocateCreditNoteNumber(ctx, tx, userID, year)
}

// FormatCreditNoteNumberForTest exposes formatCreditNoteNumber to the tests package.
func FormatCreditNoteNumberForTest(year, seq int) string {
	return formatCreditNoteNumber(year, seq)
}
