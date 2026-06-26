package actions

import (
	"context"

	"project-devis-invoice/pdp"
)

// ReportScopeClauseForTest exposes reportScopeClause to the tests package.
func ReportScopeClauseForTest(kind pdp.ReportKind, alias string) (string, bool) {
	return reportScopeClause(kind, alias)
}

// ReportingAggregateForTest exposes reportingAggregate to the tests package.
func (s *Server) ReportingAggregateForTest(ctx context.Context, userID string, kind pdp.ReportKind, period pdp.ReportPeriod) ([]pdp.ReportLine, int64, int64, error) {
	return s.reportingAggregate(ctx, userID, kind, period)
}
