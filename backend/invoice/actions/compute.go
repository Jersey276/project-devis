package actions

import (
	"math"
	"sort"
	"strconv"
)

type computeLine struct {
	ht        int64
	taxID     int32
	taxRate   float64
	taxRateID string
	taxLabel  string
}

type vatBucket struct {
	rate   string
	baseHT int64
	vat    int64
}

type computeResult struct {
	totalHT   int64
	totalVAT  int64
	totalTTC  int64
	breakdown []vatBucket
}

func computeTotals(lines []computeLine, vatExempt bool) computeResult {
	var res computeResult

	type agg struct {
		rate    string
		numRate float64
		baseHT  int64
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

func roundHalfUp(v float64) int64 {
	return int64(math.Round(v))
}

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
