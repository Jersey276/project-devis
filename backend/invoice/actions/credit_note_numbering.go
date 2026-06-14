package actions

import (
	"context"
	"database/sql"
	"fmt"
)

// allocateCreditNoteNumber consumes the next credit-note sequence value for
// (userID, year) inside the given transaction. Same gap-free guarantee as the
// invoice sequence: the number is consumed only on commit, so a rollback leaves
// the counter untouched.
func allocateCreditNoteNumber(ctx context.Context, tx *sql.Tx, userID string, year int) (number string, seq int, err error) {
	err = tx.QueryRowContext(ctx,
		`INSERT INTO credit_note_number_sequences (user_id, year, last_value)
		 VALUES ($1, $2, 1)
		 ON CONFLICT (user_id, year)
		 DO UPDATE SET last_value = credit_note_number_sequences.last_value + 1
		 RETURNING last_value`,
		userID, year,
	).Scan(&seq)
	if err != nil {
		return "", 0, fmt.Errorf("allocate credit note number: %w", err)
	}
	return formatCreditNoteNumber(year, seq), seq, nil
}

// formatCreditNoteNumber renders the legal credit-note number as "AV-YYYY-NNNN".
func formatCreditNoteNumber(year, seq int) string {
	return fmt.Sprintf("AV-%04d-%04d", year, seq)
}
