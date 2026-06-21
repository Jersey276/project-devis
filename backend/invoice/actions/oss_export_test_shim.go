package actions

import (
	"context"
	"time"

	usersGrpc "project-devis-invoice/services/usersgrpc"
)

func PickDestinationTaxForTest(taxes []*usersGrpc.Tax) (rate, label string, ok bool) {
	t := pickDestinationTax(taxes)
	if t == nil {
		return "", "", false
	}
	return t.GetRate(), t.GetName(), true
}

const OSSThresholdCentsForTest = ossThresholdCents

func OSSAppliesForTest(ossEnabled bool, cumulativeHTCents int64, clientType string, c *usersGrpc.Country) bool {
	return ossApplies(ossEnabled, cumulativeHTCents, clientType, c)
}

func (s *Server) OSSCumulativeHTForYearForTest(ctx context.Context, userID, excludeInvoiceID string, at time.Time) (int64, error) {
	return s.ossCumulativeHTForYear(ctx, userID, excludeInvoiceID, at)
}
