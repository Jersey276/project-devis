package format

import (
	"fmt"
	"strconv"
	"strings"
	"time"
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

// Date formats an RFC3339 timestamp or a "2006-01-02" date string as a
// French-locale date, e.g. "2024-03-15T10:30:00Z" → "15/03/2024".
// Returns the input unchanged if it cannot be parsed.
func Date(value string) string {
	if value == "" {
		return ""
	}
	if i := strings.IndexByte(value, 'T'); i > 0 {
		value = value[:i]
	}
	t, err := time.Parse("2006-01-02", value)
	if err != nil {
		return value
	}
	return t.Format("02/01/2006")
}

// ShortID returns the first 8 characters of an ID string,
// or the full string if shorter than 8 characters.
func ShortID(id string) string {
	if len(id) >= 8 {
		return id[:8]
	}
	return id
}
