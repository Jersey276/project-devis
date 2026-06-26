package actions

import (
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestRunPurge_Success(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(`DELETE FROM activity_logs`).
		WillReturnResult(sqlmock.NewResult(0, 3))

	n, err := runPurge(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 3 {
		t.Fatalf("expected 3 rows deleted, got %d", n)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestRunPurge_NoRows(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(`DELETE FROM activity_logs`).
		WillReturnResult(sqlmock.NewResult(0, 0))

	n, err := runPurge(db)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if n != 0 {
		t.Fatalf("expected 0 rows deleted, got %d", n)
	}
}

func TestRunPurge_DBError(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	mock.ExpectExec(`DELETE FROM activity_logs`).
		WillReturnError(errDBDown)

	_, err = runPurge(db)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
