package actions

import (
	"testing"

	quoteGrpc "project-devis-schedule/services/quotegrpc"
)

func TestQuoteLineExpectedCents_SimpleLine(t *testing.T) {
	line := &quoteGrpc.QuoteLine{
		Type:      "simple",
		Quantity:  "2",
		UnitPrice: 1000,
	}
	got, err := quoteLineExpectedCents(line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 2000 {
		t.Fatalf("expected 2000, got %d", got)
	}
}

func TestQuoteLineExpectedCents_DetailedLineSumsSublines(t *testing.T) {
	line := &quoteGrpc.QuoteLine{
		Type: "multiple",
		Data: `{"kind":"detailed","sublines":[{"quantity":"2","unit_price":1000},{"quantity":"1","unit_price":500}]}`,
	}
	got, err := quoteLineExpectedCents(line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 2500 {
		t.Fatalf("expected 2500 (sum of sublines), got %d", got)
	}
}

func TestQuoteLineExpectedCents_DetailedLineIgnoresOwnQuantity(t *testing.T) {
	// A detailed line's own quantity/unit_price are not meaningful — the
	// amount must come exclusively from sublines, never fall back to 0
	// because the parent line has no price of its own.
	line := &quoteGrpc.QuoteLine{
		Type:      "multiple",
		Quantity:  "",
		UnitPrice: 0,
		Data:      `{"kind":"detailed","sublines":[{"quantity":"3","unit_price":700}]}`,
	}
	got, err := quoteLineExpectedCents(line)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got != 2100 {
		t.Fatalf("expected 2100, got %d", got)
	}
}
