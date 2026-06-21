package actions

import (
	"context"
	"database/sql"
)

func AllocateInvoiceNumberForTest(ctx context.Context, tx *sql.Tx, userID string, year int) (string, int, error) {
	return allocateInvoiceNumber(ctx, tx, userID, year)
}

func FormatInvoiceNumberForTest(year, seq int) string {
	return formatInvoiceNumber(year, seq)
}
