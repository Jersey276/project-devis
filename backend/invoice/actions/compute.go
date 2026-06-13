package actions

import (
	"math"
	"sort"
	"strconv"
)

// computeLine is one already-resolved billable line: its frozen HT amount (in
// cents) and the VAT rate that applies to it. The HT amount is whatever is
// actually being billed — for a schedule invoice it is the sum of the selected
// months' cells; for a quote invoice it is the line's full HT.
type computeLine struct {
	ht        int64   // line HT in cents
	taxID     int32   // 0 = no tax
	taxRate   float64 // percentage, e.g. 20 for 20%
	taxRateID string  // canonical rate string used as the breakdown key, e.g. "20"
	taxLabel  string
}

// vatBucket is the aggregated VAT figure for one rate.
type vatBucket struct {
	rate    string // canonical rate string, e.g. "5.5"
	baseHT  int64
	vat     int64
}

// computeResult holds the totals and the per-rate VAT breakdown.
type computeResult struct {
	totalHT  int64
	totalVAT int64
	totalTTC int64
	breakdown []vatBucket // ordered by ascending numeric rate
}

// computeTotals aggregates HT per VAT rate, rounds the VAT once per rate (FR
// convention, consistent with the printed VAT breakdown), and sums the totals.
// When vatExempt is true no VAT is computed (art. 293 B CGI franchise) and the
// breakdown is empty.
func computeTotals(lines []computeLine, vatExempt bool) computeResult {
	var res computeResult

	// Aggregate HT per rate first; VAT is rounded once per group afterwards so
	// the figure matches the printed per-rate breakdown.
	type agg struct {
		rate   string
		numRate float64
		baseHT int64
	}
	buckets := make(map[string]*agg)
	for _, l := range lines {
		res.totalHT += l.ht
		if vatExempt {
			continue
		}
		b, ok := buckets[l.taxRateID]
		if !ok {
			b = &agg{rate: l.taxRateID, numRate: l.taxRate}
			buckets[l.taxRateID] = b
		}
		b.baseHT += l.ht
	}

	if vatExempt {
		res.totalTTC = res.totalHT
		return res
	}

	ordered := make([]*agg, 0, len(buckets))
	for _, b := range buckets {
		ordered = append(ordered, b)
	}
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].numRate != ordered[j].numRate {
			return ordered[i].numRate < ordered[j].numRate
		}
		return ordered[i].rate < ordered[j].rate
	})

	for _, b := range ordered {
		vat := roundHalfUp(float64(b.baseHT) * b.numRate / 100.0)
		res.totalVAT += vat
		res.breakdown = append(res.breakdown, vatBucket{rate: b.rate, baseHT: b.baseHT, vat: vat})
	}
	res.totalTTC = res.totalHT + res.totalVAT
	return res
}

// roundHalfUp rounds to the nearest cent, halves away from zero — the standard
// commercial rounding used on French invoices.
func roundHalfUp(v float64) int64 {
	return int64(math.Round(v))
}

// parseRate parses a tax rate string (e.g. "5.5", "20", "20.00") into a float
// percentage. Returns 0 for empty/invalid input (treated as no VAT).
func parseRate(s string) float64 {
	if s == "" {
		return 0
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return v
}
