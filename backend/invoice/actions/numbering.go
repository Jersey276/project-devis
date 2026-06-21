package actions

import (
	"context"
	"database/sql"
	"fmt"
)

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

func formatInvoiceNumber(year, seq int) string {
	return fmt.Sprintf("%04d-%04d", year, seq)
}
