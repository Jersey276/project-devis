package format

import (
	"fmt"
	"strconv"
	"strings"
)

// Cents formats a cent-denominated integer as a French-locale euro string,
// e.g. 150050 → "1 500,50 €", -500 → "-5,00 €".
func Cents(cents int64) string {
	neg := cents < 0
	if neg {
		cents = -cents
	}
	euros := cents / 100
	rem := cents % 100
	euroStr := groupThousands(strconv.FormatInt(euros, 10))
	sign := ""
	if neg {
		sign = "-"
	}
	return fmt.Sprintf("%s%s,%02d €", sign, euroStr, rem)
}

func groupThousands(s string) string {
	n := len(s)
	if n <= 3 {
		return s
	}
	var b strings.Builder
	pre := n % 3
	if pre > 0 {
		b.WriteString(s[:pre])
		if n > pre {
			b.WriteByte(' ')
		}
	}
	for i := pre; i < n; i += 3 {
		b.WriteString(s[i : i+3])
		if i+3 < n {
			b.WriteByte(' ')
		}
	}
	return b.String()
}

// Rate formats a tax-rate string as a percentage label, e.g. "20" → "20 %".
// An empty rate returns "0 %".
func Rate(rate string) string {
	rate = strings.TrimSpace(rate)
	if rate == "" {
		return "0 %"
	}
	return rate + " %"
}

// Date extracts the date portion of an RFC3339 string for display.
// "2024-03-15T..." → "2024-03-15". Returns the input unchanged if no 'T' is found.
func Date(rfc3339 string) string {
	if rfc3339 == "" {
		return ""
	}
	if i := strings.IndexByte(rfc3339, 'T'); i > 0 {
		return rfc3339[:i]
	}
	return rfc3339
}

// ShortID returns the first 8 characters of an ID string,
// or the full string if shorter than 8 characters.
func ShortID(id string) string {
	if len(id) >= 8 {
		return id[:8]
	}
	return id
}
