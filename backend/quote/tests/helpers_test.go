package tests

import (
	"testing"

	"project-devis-quote/actions"

	"github.com/DATA-DOG/go-sqlmock"
)

func setupServer(t *testing.T) (*actions.Server, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return actions.NewServer(db), mock
}

// expectEditableCheck stubs the EditableForUser pre-check.
// state="" means quote not found; otherwise the row is returned with the given
// state and a nil archived_at.
func expectEditableCheck(mock sqlmock.Sqlmock, quoteID, userID, state string) {
	rows := sqlmock.NewRows([]string{"state", "archived_at"})
	if state != "" {
		rows.AddRow(state, nil)
	}
	mock.ExpectQuery(`SELECT state, archived_at FROM quotes WHERE quote_id=\$1 AND user_id=\$2`).
		WithArgs(quoteID, userID).
		WillReturnRows(rows)
}

// expectLineParentEditable stubs LineParentEditable (JOIN of quote_lines + quotes).
// state="" means line/quote not found.
func expectLineParentEditable(mock sqlmock.Sqlmock, lineID, userID, state string) {
	rows := sqlmock.NewRows([]string{"state", "archived_at"})
	if state != "" {
		rows.AddRow(state, nil)
	}
	mock.ExpectQuery(`SELECT q\.state, q\.archived_at\s+FROM quote_lines l\s+JOIN quotes q`).
		WithArgs(lineID, userID).
		WillReturnRows(rows)
}
