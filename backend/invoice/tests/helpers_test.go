package tests

import (
	"database/sql"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

// newMockDB returns a sqlmock-backed *sql.DB for DB-level unit tests
// (numbering, sources query helpers). Downstream gRPC clients are not needed
// for these, so they are wired as nil via the server's constructor elsewhere.
func newMockDB(t *testing.T) (*sql.DB, sqlmock.Sqlmock) {
	t.Helper()
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	t.Cleanup(func() { db.Close() })
	return db, mock
}
