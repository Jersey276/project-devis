package actions

import (
	"context"
	"database/sql"
)

// AllocateInvoiceNumberForTest exposes allocateInvoiceNumber to the external
// tests package. Thin adapter, no logic of its own.
func AllocateInvoiceNumberForTest(ctx context.Context, tx *sql.Tx, userID string, year int) (string, int, error) {
	return allocateInvoiceNumber(ctx, tx, userID, year)
}

// FormatInvoiceNumberForTest exposes formatInvoiceNumber to the tests package.
func FormatInvoiceNumberForTest(year, seq int) string {
	return formatInvoiceNumber(year, seq)
}
