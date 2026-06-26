package actions

import (
	"context"
	"database/sql"
)

func AllocateCreditNoteNumberForTest(ctx context.Context, tx *sql.Tx, userID string, year int) (string, int, error) {
	return allocateCreditNoteNumber(ctx, tx, userID, year)
}

func FormatCreditNoteNumberForTest(year, seq int) string {
	return formatCreditNoteNumber(year, seq)
}
