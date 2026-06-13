package tests

import (
	"testing"

	"project-devis-invoice/actions"
)

// These tests pin the fiscal contract of the VAT computation. They go through
// the test-only export shim actions.ComputeTotalsForTest so they can live in
// the tests package while compute itself stays unexported.

func TestComputeTotals_SingleRate20(t *testing.T) {
	// 100.00 € HT at 20% → 20.00 € VAT → 120.00 € TTC
	lines := []actions.ComputeLineInput{
		{HT: 10000, TaxRate: 20, TaxRateID: "20"},
	}
	got := actions.ComputeTotalsForTest(lines, false)

	if got.TotalHT != 10000 || got.TotalVAT != 2000 || got.TotalTTC != 12000 {
		t.Fatalf("totals = HT %d / VAT %d / TTC %d; want 10000 / 2000 / 12000",
			got.TotalHT, got.TotalVAT, got.TotalTTC)
	}
	if len(got.Breakdown) != 1 {
		t.Fatalf("breakdown len = %d; want 1", len(got.Breakdown))
	}
	if b := got.Breakdown[0]; b.Rate != "20" || b.BaseHT != 10000 || b.VAT != 2000 {
		t.Fatalf("breakdown[0] = %+v; want {20 10000 2000}", b)
	}
}

func TestComputeTotals_MultiRateOrderedAscending(t *testing.T) {
	// Mix of 20%, 10%, 5.5% and 0% — breakdown must be ordered by ascending
	// numeric rate and each VAT rounded once per rate group.
	lines := []actions.ComputeLineInput{
		{HT: 10000, TaxRate: 20, TaxRateID: "20"},    // 2000 VAT
		{HT: 5000, TaxRate: 10, TaxRateID: "10"},     // 500 VAT
		{HT: 3333, TaxRate: 5.5, TaxRateID: "5.5"},   // 183.315 → 183 VAT
		{HT: 4000, TaxRate: 0, TaxRateID: "0"},       // 0 VAT
	}
	got := actions.ComputeTotalsForTest(lines, false)

	wantHT := int64(10000 + 5000 + 3333 + 4000)
	wantVAT := int64(2000 + 500 + 183 + 0)
	if got.TotalHT != wantHT || got.TotalVAT != wantVAT || got.TotalTTC != wantHT+wantVAT {
		t.Fatalf("totals = HT %d / VAT %d / TTC %d; want %d / %d / %d",
			got.TotalHT, got.TotalVAT, got.TotalTTC, wantHT, wantVAT, wantHT+wantVAT)
	}

	wantRates := []string{"0", "5.5", "10", "20"}
	if len(got.Breakdown) != len(wantRates) {
		t.Fatalf("breakdown len = %d; want %d", len(got.Breakdown), len(wantRates))
	}
	for i, want := range wantRates {
		if got.Breakdown[i].Rate != want {
			t.Fatalf("breakdown[%d].Rate = %q; want %q (order must be ascending)",
				i, got.Breakdown[i].Rate, want)
		}
	}
}

func TestComputeTotals_RoundsPerRateGroupNotPerLine(t *testing.T) {
	// Two lines at the same 5.5% rate. Per-line rounding would give
	// round(50.005)+round(50.005) = ... whereas per-group rounding sums the HT
	// first: (1001 + 1001) * 5.5% = 2002 * 0.055 = 110.11 → 110.
	// Per-line: round(1001*0.055)=round(55.055)=55 twice = 110 as well here, so
	// pick figures where the two differ: HT 909 + 909 = 1818.
	// 909 * 0.055 = 49.995 → round 50 each → 100 per line.
	// group: 1818 * 0.055 = 99.99 → 100. Use HT that splits: 901 and 901.
	// 901*0.055=49.555→50 each→100; group 1802*0.055=99.11→99. They differ.
	lines := []actions.ComputeLineInput{
		{HT: 901, TaxRate: 5.5, TaxRateID: "5.5"},
		{HT: 901, TaxRate: 5.5, TaxRateID: "5.5"},
	}
	got := actions.ComputeTotalsForTest(lines, false)

	if len(got.Breakdown) != 1 {
		t.Fatalf("breakdown len = %d; want 1 (same rate must aggregate)", len(got.Breakdown))
	}
	if got.Breakdown[0].BaseHT != 1802 {
		t.Fatalf("base HT = %d; want 1802 (HT must aggregate before rounding)", got.Breakdown[0].BaseHT)
	}
	// round(1802 * 0.055) = round(99.11) = 99
	if got.TotalVAT != 99 {
		t.Fatalf("total VAT = %d; want 99 (per-group rounding)", got.TotalVAT)
	}
}

func TestComputeTotals_VATExempt(t *testing.T) {
	lines := []actions.ComputeLineInput{
		{HT: 10000, TaxRate: 20, TaxRateID: "20"},
		{HT: 5000, TaxRate: 10, TaxRateID: "10"},
	}
	got := actions.ComputeTotalsForTest(lines, true)

	if got.TotalHT != 15000 {
		t.Fatalf("total HT = %d; want 15000", got.TotalHT)
	}
	if got.TotalVAT != 0 || got.TotalTTC != 15000 {
		t.Fatalf("exempt totals = VAT %d / TTC %d; want 0 / 15000", got.TotalVAT, got.TotalTTC)
	}
	if len(got.Breakdown) != 0 {
		t.Fatalf("exempt breakdown len = %d; want 0", len(got.Breakdown))
	}
}

func TestComputeTotals_Empty(t *testing.T) {
	got := actions.ComputeTotalsForTest(nil, false)
	if got.TotalHT != 0 || got.TotalVAT != 0 || got.TotalTTC != 0 || len(got.Breakdown) != 0 {
		t.Fatalf("empty result = %+v; want all zero", got)
	}
}
