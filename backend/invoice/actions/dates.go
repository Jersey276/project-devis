package actions

import "time"

const defaultDueInDays = 30

var invoiceTZ = mustLoadParis()

func mustLoadParis() *time.Location {
	loc, err := time.LoadLocation("Europe/Paris")
	if err != nil {
		return time.UTC
	}
	return loc
}

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
