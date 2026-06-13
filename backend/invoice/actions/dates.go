package actions

import "time"

const defaultDueInDays = 30

// invoiceTZ is the timezone that fixes the legal issue date (and therefore the
// numbering year). French invoices use Europe/Paris.
var invoiceTZ = mustLoadParis()

func mustLoadParis() *time.Location {
	loc, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		return time.UTC // fallback if tzdata is unavailable in the image
	}
	return loc
}

// resolveSaleAndDue derives the sale and due dates from the request. saleDate is
// "YYYY-MM-DD" or empty (defaults to the issue date). dueInDays defaults to 30.
func resolveSaleAndDue(issuedAt time.Time, saleDate string, dueInDays int32) (sale, due time.Time) {
	sale = issuedAt
	if saleDate != "" {
		if parsed, err := time.ParseInLocation("2006-01-02", saleDate, invoiceTZ); err == nil {
			sale = parsed
		}
	}
	days := int(dueInDays)
	if days <= 0 {
		days = defaultDueInDays
	}
	due = issuedAt.AddDate(0, 0, days)
	return sale, due
}
