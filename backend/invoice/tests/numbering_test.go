package tests

import (
	"context"
	"testing"

	"project-devis-invoice/actions"

	"github.com/DATA-DOG/go-sqlmock"
)

func TestFormatInvoiceNumber(t *testing.T) {
	cases := []struct {
		year, seq int
		want      string
	}{
		{2026, 1, "2026-0001"},
		{2026, 42, "2026-0042"},
		{2026, 9999, "2026-9999"},
		{2027, 1, "2027-0001"},
	}
	for _, c := range cases {
		if got := actions.FormatInvoiceNumberForTest(c.year, c.seq); got != c.want {
			t.Errorf("FormatInvoiceNumber(%d,%d) = %q; want %q", c.year, c.seq, got, c.want)
		}
	}
}

func TestAllocateInvoiceNumber_FirstOfYear(t *testing.T) {
	db, mock := newMockDB(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO invoice_number_sequences .* ON CONFLICT .* DO UPDATE SET last_value = invoice_number_sequences.last_value \+ 1\s+RETURNING last_value`).
		WithArgs("user-1", 2026).
		WillReturnRows(sqlmock.NewRows([]string{"last_value"}).AddRow(1))

	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	number, seq, err := actions.AllocateInvoiceNumberForTest(context.Background(), tx, "user-1", 2026)
	if err != nil {
		t.Fatalf("allocate: %v", err)
	}
	if number != "2026-0001" || seq != 1 {
		t.Fatalf("got number=%q seq=%d; want 2026-0001 / 1", number, seq)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}

func TestAllocateInvoiceNumber_IncrementsWithinYear(t *testing.T) {
	db, mock := newMockDB(t)

	mock.ExpectBegin()
	mock.ExpectQuery(`INSERT INTO invoice_number_sequences`).
		WithArgs("user-1", 2026).
		WillReturnRows(sqlmock.NewRows([]string{"last_value"}).AddRow(7))

	tx, _ := db.Begin()
	number, seq, err := actions.AllocateInvoiceNumberForTest(context.Background(), tx, "user-1", 2026)
	if err != nil {
		t.Fatalf("allocate: %v", err)
	}
	if number != "2026-0007" || seq != 7 {
		t.Fatalf("got number=%q seq=%d; want 2026-0007 / 7", number, seq)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet expectations: %v", err)
	}
}
