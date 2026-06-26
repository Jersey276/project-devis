package facturx

import "testing"

func TestAmountFromCents(t *testing.T) {
	cases := map[int64]string{
		0:       "0.00",
		5:       "0.05",
		99:      "0.99",
		100:     "1.00",
		12345:   "123.45",
		-12345:  "-123.45",
		1000000: "10000.00",
	}
	for in, want := range cases {
		if got := amountFromCents(in); got != want {
			t.Errorf("amountFromCents(%d) = %q; want %q", in, got, want)
		}
	}
}

func TestPercentFromRate(t *testing.T) {
	cases := map[string]string{
		"20":  "20.00",
		"5.5": "5.50",
		"5,5": "5.50",
		"0":   "0.00",
		"":    "0.00",
		"x":   "0.00",
	}
	for in, want := range cases {
		if got := percentFromRate(in); got != want {
			t.Errorf("percentFromRate(%q) = %q; want %q", in, got, want)
		}
	}
}

func TestDateCII(t *testing.T) {
	cases := map[string]string{
		"2026-06-14T10:30:00Z": "20260614",
		"2026-06-14":           "20260614",
		"":                     "",
	}
	for in, want := range cases {
		if got := dateCII(in); got != want {
			t.Errorf("dateCII(%q) = %q; want %q", in, got, want)
		}
	}
}

func TestCategoryForRate(t *testing.T) {
	if got := categoryForRate("20", true); got != categoryNotSubject {
		t.Errorf("franchise (293 B) invoice should be category O, got %q", got)
	}
	if got := categoryForRate("0", false); got != categoryZero {
		t.Errorf("zero rate should be category Z, got %q", got)
	}
	if got := categoryForRate("20", false); got != categoryStandard {
		t.Errorf("standard rate should be category S, got %q", got)
	}
}
