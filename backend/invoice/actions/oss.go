package actions

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	usersGrpc "project-devis-invoice/services/usersgrpc"
)

const ossThresholdCents int64 = 1_000_000

func isIntraEUB2C(clientType string, c *usersGrpc.Country) bool {
	if c == nil || clientType != "individual" {
		return false
	}
	return c.GetIsEu() && c.GetCode() != "FR"
}

func ossApplies(ossEnabled bool, cumulativeHTCents int64, clientType string, c *usersGrpc.Country) bool {
	if !isIntraEUB2C(clientType, c) {
		return false
	}
	return ossEnabled || cumulativeHTCents >= ossThresholdCents
}

func (s *Server) ossCumulativeHTForYear(ctx context.Context, userID, excludeInvoiceID string, at time.Time) (int64, error) {
	y := at.In(invoiceTZ).Year()
	start := time.Date(y, 1, 1, 0, 0, 0, 0, invoiceTZ)
	end := start.AddDate(1, 0, 0)

	// Net assiette = issued invoices in the OSS scope, minus credit notes that
	// neutralise part of them (also in scope). Crediting reduces the distance-sale
	// turnover that counts toward the EUR 10 000 threshold (art. 259 D CGI), so the
	// switch is driven by the actual net sales, not the gross. Both legs are frozen
	// at issue time via counts_toward_oss_threshold (see ADR 0002).
	var invoicesHT sql.NullInt64
	err := s.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(i.total_ht_cents), 0)
		   FROM invoices i
		   JOIN invoice_party_snapshots p ON p.invoice_id = i.invoice_id
		  WHERE i.user_id = $1
		    AND i.invoice_id <> $2
		    AND i.status IN ('ISSUED', 'PAID')
		    AND p.counts_toward_oss_threshold
		    AND i.issued_at >= $3 AND i.issued_at < $4`,
		userID, excludeInvoiceID, start, end,
	).Scan(&invoicesHT)
	if err != nil {
		return 0, fmt.Errorf("oss cumulative invoices: %w", err)
	}

	var creditNotesHT sql.NullInt64
	err = s.db.QueryRowContext(ctx,
		`SELECT COALESCE(SUM(cn.total_ht_cents), 0)
		   FROM credit_notes cn
		   JOIN credit_note_party_snapshots p ON p.credit_note_id = cn.credit_note_id
		  WHERE cn.user_id = $1
		    AND p.counts_toward_oss_threshold
		    AND cn.issued_at >= $2 AND cn.issued_at < $3`,
		userID, start, end,
	).Scan(&creditNotesHT)
	if err != nil {
		return 0, fmt.Errorf("oss cumulative credit notes: %w", err)
	}

	return invoicesHT.Int64 - creditNotesHT.Int64, nil
}
