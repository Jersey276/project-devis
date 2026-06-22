package pdp

import "context"

// ReportKind is the category of e-reporting transmitted to the platform, distinct
// from B6 invoice deposit: an aggregate of out-of-scope-of-e-invoicing operations
// for a period. TRANSACTION = domestic B2C sales (B5); CROSS_BORDER_B2C = intra-EU
// distance sales, the OSS scope (C5).
type ReportKind string

const (
	ReportTransaction    ReportKind = "TRANSACTION"
	ReportCrossBorderB2C ReportKind = "CROSS_BORDER_B2C"
)

// ReportPeriod is the civil month (Europe/Paris) an e-report covers.
type ReportPeriod struct {
	Year  int
	Month int // 1-12
}

// ReportLine is one VAT-rate (and, for cross-border, destination-country) bucket
// of the period aggregate. Amounts are net of credit notes, in cents.
type ReportLine struct {
	TaxRate     string
	CountryCode string // destination country; "FR" for domestic transaction reports
	BaseHTCents int64
	VATCents    int64
}

// SubmitReportInput carries one period aggregate to transmit. Kept declarative
// (already-aggregated lines + totals) so the seam holds no business logic.
type SubmitReportInput struct {
	UserID        string
	Kind          ReportKind
	Period        ReportPeriod
	Lines         []ReportLine
	TotalHTCents  int64
	TotalVATCents int64
}

// SubmitReportResult is the outcome of transmitting one period aggregate.
type SubmitReportResult struct {
	ReportID string         // platform-assigned handle; "" for the no-op adapter
	Status   PlatformStatus // typically PlatformSubmitted on success
}

// Reporter is the neutral seam for French e-reporting (B5/C5): periodic
// transmission of transaction/cross-border aggregates to a Plateforme Agréée.
// Like Client (pdp.go) it carries no proto and no DB; the default adapter is a
// no-op, so the feature is inert until a real PA is wired.
type Reporter interface {
	// SubmitReport transmits one period aggregate and returns the platform handle.
	SubmitReport(ctx context.Context, in SubmitReportInput) (SubmitReportResult, error)
	// FetchReportStatus polls the current platform status for a prior report.
	FetchReportStatus(ctx context.Context, reportID string) (PlatformStatus, error)
}
