package actions

type ComputeLineInput struct {
	HT        int64
	TaxID     int32
	TaxRate   float64
	TaxRateID string
	TaxLabel  string
}

type ComputeVATLineOutput struct {
	Rate   string
	BaseHT int64
	VAT    int64
}

type ComputeResultOutput struct {
	TotalHT   int64
	TotalVAT  int64
	TotalTTC  int64
	Breakdown []ComputeVATLineOutput
}

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
