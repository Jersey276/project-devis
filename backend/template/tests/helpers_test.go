package tests

import (
	"testing"

	"project-devis-template/actions"

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
