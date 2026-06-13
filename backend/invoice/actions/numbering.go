package actions

import (
	"context"
	"database/sql"
	"fmt"
)

// allocateInvoiceNumber consumes the next sequence value for (userID, year)
// inside the given transaction and returns the formatted invoice number plus
// its raw components. The counter is incremented atomically via the row lock
// taken by the upsert; because it runs inside the issue transaction, a rollback
// leaves the counter untouched — guaranteeing a gap-free sequence (art. 289 CGI).
//
func allocateInvoiceNumber(ctx context.Context, tx *sql.Tx, userID string, year int) (number string, seq int, err error) {
	err = tx.QueryRowContext(ctx,
		`INSERT INTO invoice_number_sequences (user_id, year, last_value)
		 VALUES ($1, $2, 1)
		 ON CONFLICT (user_id, year)
		 DO UPDATE SET last_value = invoice_number_sequences.last_value + 1
		 RETURNING last_value`,
		userID, year,
	).Scan(&seq)
	if err != nil {
		return "", 0, fmt.Errorf("allocate invoice number: %w", err)
	}
	return formatInvoiceNumber(year, seq), seq, nil
}

// formatInvoiceNumber renders the legal invoice number as "YYYY-NNNN".
func formatInvoiceNumber(year, seq int) string {
	return fmt.Sprintf("%04d-%04d", year, seq)
}
