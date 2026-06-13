package actions

// This file exposes the otherwise-unexported VAT computation to the external
// `tests` package, mirroring the test-facing shims used elsewhere (e.g.
// schedule's SetQuoteLineExpectedCentsFetcherForTests). It contains no business
// logic — only a thin adapter over computeTotals.

// ComputeLineInput is the test-facing shape of a billable line.
type ComputeLineInput struct {
	HT        int64
	TaxID     int32
	TaxRate   float64
	TaxRateID string
	TaxLabel  string
}

// ComputeVATLineOutput is the test-facing per-rate VAT figure.
type ComputeVATLineOutput struct {
	Rate   string
	BaseHT int64
	VAT    int64
}

// ComputeResultOutput is the test-facing computation result.
type ComputeResultOutput struct {
	TotalHT   int64
	TotalVAT  int64
	TotalTTC  int64
	Breakdown []ComputeVATLineOutput
}

// ComputeTotalsForTest adapts ComputeLineInput → computeLine, runs the real
// computeTotals, and adapts the result back to exported types.
func ComputeTotalsForTest(lines []ComputeLineInput, vatExempt bool) ComputeResultOutput {
	in := make([]computeLine, len(lines))
	for i, l := range lines {
		in[i] = computeLine{
			ht:        l.HT,
			taxID:     l.TaxID,
			taxRate:   l.TaxRate,
			taxRateID: l.TaxRateID,
			taxLabel:  l.TaxLabel,
		}
	}
	res := computeTotals(in, vatExempt)
	out := ComputeResultOutput{
		TotalHT:  res.totalHT,
		TotalVAT: res.totalVAT,
		TotalTTC: res.totalTTC,
	}
	for _, b := range res.breakdown {
		out.Breakdown = append(out.Breakdown, ComputeVATLineOutput{Rate: b.rate, BaseHT: b.baseHT, VAT: b.vat})
	}
	return out
}
