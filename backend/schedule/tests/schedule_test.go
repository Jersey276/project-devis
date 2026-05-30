package tests

import (
	"testing"

	"project-devis-schedule/actions"
)

func TestNewServer(t *testing.T) {
	srv, mock := setupServer(t)

	if srv == nil {
		t.Fatal("expected non-nil server")
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unexpected DB calls: %v", err)
	}
}

func TestValidateCreateScheduleInput(t *testing.T) {
	tests := []struct {
		name          string
		userID        string
		quoteID       string
		scheduleName  string
		startMonth    string
		durationMonths int32
		wantErr       bool
	}{
		{name: "valid input", userID: "user-1", quoteID: "quote-1", scheduleName: "Echeancier principal", startMonth: "2026-06", durationMonths: 12},
		{name: "missing user", quoteID: "quote-1", scheduleName: "Echeancier principal", startMonth: "2026-06", durationMonths: 12, wantErr: true},
		{name: "missing quote", userID: "user-1", scheduleName: "Echeancier principal", startMonth: "2026-06", durationMonths: 12, wantErr: true},
		{name: "missing name", userID: "user-1", quoteID: "quote-1", startMonth: "2026-06", durationMonths: 12, wantErr: true},
		{name: "invalid month format", userID: "user-1", quoteID: "quote-1", scheduleName: "Echeancier principal", startMonth: "06-2026", durationMonths: 12, wantErr: true},
		{name: "invalid month value", userID: "user-1", quoteID: "quote-1", scheduleName: "Echeancier principal", startMonth: "2026-13", durationMonths: 12, wantErr: true},
		{name: "zero duration", userID: "user-1", quoteID: "quote-1", scheduleName: "Echeancier principal", startMonth: "2026-06", durationMonths: 0, wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := actions.ValidateCreateScheduleInput(tc.userID, tc.quoteID, tc.scheduleName, tc.startMonth, tc.durationMonths)
			if tc.wantErr && err == nil {
				t.Fatal("expected validation error")
			}
			if !tc.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestParseAmountEuros(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		wantCents int64
		wantErr   bool
	}{
		{name: "zero", input: "0", wantCents: 0},
		{name: "integer euros", input: "12", wantCents: 1200},
		{name: "two decimals", input: "12.30", wantCents: 1230},
		{name: "negative rejected", input: "-1.00", wantErr: true},
		{name: "too many decimals rejected", input: "12.345", wantErr: true},
		{name: "invalid number rejected", input: "abc", wantErr: true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := actions.ParseAmountEuros(tc.input)
			if tc.wantErr {
				if err == nil {
					t.Fatal("expected error")
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.wantCents {
				t.Fatalf("expected %d cents, got %d", tc.wantCents, got)
			}
		})
	}
}

func TestIsEditableStatus(t *testing.T) {
	tests := []struct {
		status string
		want   bool
	}{
		{status: actions.StatusDraft, want: true},
		{status: actions.StatusNegotiate, want: true},
		{status: actions.StatusDenied, want: false},
		{status: actions.StatusValid, want: false},
		{status: "weird", want: false},
	}

	for _, tc := range tests {
		if got := actions.IsEditableStatus(tc.status); got != tc.want {
			t.Fatalf("status %q: expected %v, got %v", tc.status, tc.want, got)
		}
	}
}