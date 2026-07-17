package actions

import (
	"testing"

	quoteGrpc "project-devis-invoice/services/quotegrpc"
)

func TestLineHTFromQuoteLine_SimpleLine(t *testing.T) {
	l := &quoteGrpc.QuoteLine{Type: "simple", Quantity: "2", UnitPrice: 1000}
	got := lineHTFromQuoteLine(l)
	if got != 2000 {
		t.Fatalf("expected 2000, got %d", got)
	}
}

func TestLineHTFromQuoteLine_DetailedLineSumsSublines(t *testing.T) {
	l := &quoteGrpc.QuoteLine{
		Type: "multiple",
		Data: `{"kind":"detailed","sublines":[{"quantity":"2","unit_price":1000},{"quantity":"1","unit_price":500}]}`,
	}
	got := lineHTFromQuoteLine(l)
	if got != 2500 {
		t.Fatalf("expected 2500 (sum of sublines), got %d", got)
	}
}

func TestLineQuantityAndUnitPrice_SimpleLine(t *testing.T) {
	l := &quoteGrpc.QuoteLine{Type: "simple", Quantity: "3", UnitPrice: 750}
	qty, unitPriceCents := lineQuantityAndUnitPrice(l, 2250)
	if qty != "3" {
		t.Fatalf("expected quantity 3, got %q", qty)
	}
	if unitPriceCents != 750 {
		t.Fatalf("expected unit price 750, got %d", unitPriceCents)
	}
}

func TestLineQuantityAndUnitPrice_DetailedLineDerivesFromTotal(t *testing.T) {
	// A detailed line's own quantity/unit_price are empty/zero — the snapshot
	// must not surface a misleading 0 unit price when the line total is
	// positive (e.g. billed over a subset of schedule months).
	l := &quoteGrpc.QuoteLine{
		Type: "multiple",
		Data: `{"kind":"detailed","sublines":[{"quantity":"2","unit_price":1000}]}`,
	}
	qty, unitPriceCents := lineQuantityAndUnitPrice(l, 2000)
	if qty != "1" {
		t.Fatalf("expected derived quantity 1, got %q", qty)
	}
	if unitPriceCents != 2000 {
		t.Fatalf("expected derived unit price 2000 (== line total), got %d", unitPriceCents)
	}
}

func TestQuoteLineKind_LegacyMultipleType(t *testing.T) {
	l := &quoteGrpc.QuoteLine{Type: "multiple"}
	if kind := quoteLineKind(l); kind != "detailed" {
		t.Fatalf("expected legacy type=multiple to resolve to kind=detailed, got %q", kind)
	}
}
