package actions

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"project-devis-invoice/pdp"
)

// reportScopeClause is the snapshot predicate selecting which sales belong to a
// report kind. Both scopes are B2C (client_type='individual'); they differ only on
// the destination country, so the two never overlap:
//   - TRANSACTION (B5): domestic sales (country FR).
//   - CROSS_BORDER_B2C (C5): intra-EU distance sales, the frozen OSS assiette
//     (country ≠ FR, counts_toward_oss_threshold), so it lines up exactly with the
//     OSS cumulative (see ADR 0002).
func reportScopeClause(kind pdp.ReportKind, alias string) (string, bool) {
	switch kind {
	case pdp.ReportTransaction:
		return alias + ".client_type = 'individual' AND " + alias + ".client_country_code = 'FR'", true
	case pdp.ReportCrossBorderB2C:
		return alias + ".client_type = 'individual' AND " + alias + ".client_country_code <> 'FR' AND " + alias + ".counts_toward_oss_threshold", true
	default:
		return "", false
	}
}

// reportPeriodBounds returns the [start, end) civil-month window in Europe/Paris.
func reportPeriodBounds(p pdp.ReportPeriod) (time.Time, time.Time) {
	start := time.Date(p.Year, time.Month(p.Month), 1, 0, 0, 0, 0, invoiceTZ)
	return start, start.AddDate(0, 1, 0)
}

// reportingAggregate computes one period aggregate for a report kind: net (issued
// invoices minus credit notes) HT/VAT per VAT rate and destination country, frozen
// at issue time via the party snapshot. Mirrors the OSS net assiette (ossCumulative*
// in oss.go): same ISSUED|PAID + frozen-snapshot logic, but broken down per rate so
// the platform receives a VAT breakdown rather than a single sum.
func (s *Server) reportingAggregate(ctx context.Context, userID string, kind pdp.ReportKind, period pdp.ReportPeriod) ([]pdp.ReportLine, int64, int64, error) {
	scope, ok := reportScopeClause(kind, "p")
	if !ok {
		return nil, 0, 0, fmt.Errorf("unknown report kind %q", kind)
	}
	start, end := reportPeriodBounds(period)

	// Buckets keyed by (country, rate). Credit notes net against the same bucket
	// (their snapshot carries the same frozen country/assiette flag, inherited from
	// the origin invoice), so a credited sale reduces the reported amount.
	type bucket struct {
		country, rate string
		baseHT, vat   int64
	}
	buckets := map[string]*bucket{}
	key := func(country, rate string) string { return country + "|" + rate }
	add := func(country, rate string, baseHT, vat int64) {
		k := key(country, rate)
		b := buckets[k]
		if b == nil {
			b = &bucket{country: country, rate: rate}
			buckets[k] = b
		}
		b.baseHT += baseHT
		b.vat += vat
	}

	invoiceRows, err := s.db.QueryContext(ctx,
		`SELECT p.client_country_code, v.tax_rate,
		        COALESCE(SUM(v.base_ht_cents), 0), COALESCE(SUM(v.vat_cents), 0)
		   FROM invoices i
		   JOIN invoice_party_snapshots p ON p.invoice_id = i.invoice_id
		   JOIN invoice_vat_breakdown_snapshots v ON v.invoice_id = i.invoice_id
		  WHERE i.user_id = $1
		    AND i.status IN ('ISSUED', 'PAID')
		    AND i.issued_at >= $2 AND i.issued_at < $3
		    AND (`+scope+`)
		  GROUP BY p.client_country_code, v.tax_rate`,
		userID, start, end,
	)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("report aggregate invoices: %w", err)
	}
	defer invoiceRows.Close()
	for invoiceRows.Next() {
		var country, rate string
		var baseHT, vat int64
		if err := invoiceRows.Scan(&country, &rate, &baseHT, &vat); err != nil {
			return nil, 0, 0, err
		}
		add(country, rate, baseHT, vat)
	}
	if err := invoiceRows.Err(); err != nil {
		return nil, 0, 0, err
	}

	creditRows, err := s.db.QueryContext(ctx,
		`SELECT p.client_country_code, v.tax_rate,
		        COALESCE(SUM(v.base_ht_cents), 0), COALESCE(SUM(v.vat_cents), 0)
		   FROM credit_notes cn
		   JOIN credit_note_party_snapshots p ON p.credit_note_id = cn.credit_note_id
		   JOIN credit_note_vat_breakdown_snapshots v ON v.credit_note_id = cn.credit_note_id
		  WHERE cn.user_id = $1
		    AND cn.issued_at >= $2 AND cn.issued_at < $3
		    AND (`+scope+`)
		  GROUP BY p.client_country_code, v.tax_rate`,
		userID, start, end,
	)
	if err != nil {
		return nil, 0, 0, fmt.Errorf("report aggregate credit notes: %w", err)
	}
	defer creditRows.Close()
	for creditRows.Next() {
		var country, rate string
		var baseHT, vat int64
		if err := creditRows.Scan(&country, &rate, &baseHT, &vat); err != nil {
			return nil, 0, 0, err
		}
		add(country, rate, -baseHT, -vat)
	}
	if err := creditRows.Err(); err != nil {
		return nil, 0, 0, err
	}

	lines := make([]pdp.ReportLine, 0, len(buckets))
	var totalHT, totalVAT int64
	for _, b := range buckets {
		lines = append(lines, pdp.ReportLine{
			TaxRate:     b.rate,
			CountryCode: b.country,
			BaseHTCents: b.baseHT,
			VATCents:    b.vat,
		})
		totalHT += b.baseHT
		totalVAT += b.vat
	}
	return lines, totalHT, totalVAT, nil
}

// reportKindFromString validates the wire string and returns the typed kind.
func reportKindFromString(s string) (pdp.ReportKind, bool) {
	k := pdp.ReportKind(s)
	switch k {
	case pdp.ReportTransaction, pdp.ReportCrossBorderB2C:
		return k, true
	default:
		return "", false
	}
}

// scanReportStatus reads the current report status; sql.ErrNoRows means none yet.
func scanReportStatus(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, userID string, kind pdp.ReportKind, period pdp.ReportPeriod) (string, error) {
	var status string
	err := q.QueryRowContext(ctx,
		`SELECT status FROM invoice_reports
		  WHERE user_id=$1 AND kind=$2 AND period_year=$3 AND period_month=$4`,
		userID, string(kind), period.Year, period.Month,
	).Scan(&status)
	return status, err
}
