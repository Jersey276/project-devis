package actions

import "testing"

func TestBuildOrderBy(t *testing.T) {
	allowed := map[string]string{
		"name":   "invoice_name",
		"status": "status",
	}

	cases := []struct {
		sortBy    string
		sortDir   string
		want      string
	}{
		{"name", "asc", "invoice_name ASC"},
		{"name", "ASC", "invoice_name ASC"},
		{"name", "desc", "invoice_name DESC"},
		{"status", "asc", "status ASC"},
		{"unknown", "asc", "created_at ASC"},  // whitelist fallback
		{"unknown", "desc", "created_at DESC"}, // whitelist fallback
		{"name", "", "invoice_name DESC"},      // empty direction → DESC
	}

	for _, tc := range cases {
		got := buildOrderBy(allowed, "created_at", tc.sortBy, tc.sortDir)
		if got != tc.want {
			t.Errorf("buildOrderBy(%q, %q) = %q, want %q", tc.sortBy, tc.sortDir, got, tc.want)
		}
	}
}

func TestBuildInvoiceOrderBy(t *testing.T) {
	if got := buildInvoiceOrderBy("dueDate", "asc"); got != "due_date ASC" {
		t.Errorf("got %q", got)
	}
	if got := buildInvoiceOrderBy("", "desc"); got != "created_at DESC" {
		t.Errorf("got %q", got)
	}
}

func TestBuildCreditNoteOrderBy(t *testing.T) {
	if got := buildCreditNoteOrderBy("issuedAt", "desc"); got != "cn.issued_at DESC" {
		t.Errorf("got %q", got)
	}
	if got := buildCreditNoteOrderBy("bad", "asc"); got != "cn.created_at ASC" {
		t.Errorf("got %q", got)
	}
}
