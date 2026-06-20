package actions

import (
	"context"
	"time"

	usersGrpc "project-devis-invoice/services/usersgrpc"
)

// PickDestinationTaxForTest exposes the unexported destination-rate selection
// used by OSS resolution to the external `tests` package. It returns the chosen
// tax's rate string and label, plus ok=false when no tax is available.
func PickDestinationTaxForTest(taxes []*usersGrpc.Tax) (rate, label string, ok bool) {
	t := pickDestinationTax(taxes)
	if t == nil {
		return "", "", false
	}
	return t.GetRate(), t.GetName(), true
}

// OSSThresholdCentsForTest exposes the OSS threshold constant to the external
// `tests` package.
const OSSThresholdCentsForTest = ossThresholdCents

// OSSAppliesForTest exposes the pure OSS decision to the external `tests`
// package.
func OSSAppliesForTest(ossEnabled bool, cumulativeHTCents int64, clientType string, c *usersGrpc.Country) bool {
	return ossApplies(ossEnabled, cumulativeHTCents, clientType, c)
}

// OSSCumulativeHTForYearForTest exposes the unexported yearly-cumulative query
// (DB-backed) to the external `tests` package.
func (s *Server) OSSCumulativeHTForYearForTest(ctx context.Context, userID, excludeInvoiceID string, at time.Time) (int64, error) {
	return s.ossCumulativeHTForYear(ctx, userID, excludeInvoiceID, at)
}
