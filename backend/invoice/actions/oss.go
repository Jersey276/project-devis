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

	var sum sql.NullInt64
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
	).Scan(&sum)
	if err != nil {
		return 0, fmt.Errorf("oss cumulative: %w", err)
	}
	return sum.Int64, nil
}
