package actions

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"project-devis-invoice/actions/codes"
	"project-devis-invoice/pdp"
	invoiceGrpc "project-devis-invoice/services/grpc"
)

// terminalReportStatuses are the report statuses past which a re-submission is
// refused: the period is already settled with the platform.
var terminalReportStatuses = map[string]bool{"APPROVED": true, "COLLECTED": true}

// SubmitInvoiceReport transmits one period aggregate to the platform (B5/C5) and
// records it. The aggregate is computed locally from frozen snapshots (net of
// credit notes), the platform call happens first, and only on success is the
// invoice_reports row upserted to DEPOSITED. A period already APPROVED/COLLECTED
// is not re-submitted. No-op platform by default, so this is inert in production.
func (s *Server) SubmitInvoiceReport(ctx context.Context, req *invoiceGrpc.SubmitInvoiceReportRequest) (resp *invoiceGrpc.GenericResponse, err error) {
	startedAt := time.Now()
	defer deferObserve("submit_invoice_report", startedAt, func() (int32, bool) {
		if resp == nil {
			return codes.InternalError, false
		}
		return resp.Code, resp.Success
	}, &err)()

	if req == nil || strings.TrimSpace(req.UserId) == "" {
		return &invoiceGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}
	kind, ok := reportKindFromString(req.GetKind())
	if !ok || req.GetMonth() < 1 || req.GetMonth() > 12 || req.GetYear() < 2000 {
		return &invoiceGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}
	period := pdp.ReportPeriod{Year: int(req.GetYear()), Month: int(req.GetMonth())}

	// Refuse to re-transmit a settled period.
	if current, err := scanReportStatus(ctx, s.db, req.UserId, kind, period); err == nil && terminalReportStatuses[current] {
		return &invoiceGrpc.GenericResponse{Success: false, Code: codes.LifecycleTransitionInvalid}, nil
	} else if err != nil && err != sql.ErrNoRows {
		return &invoiceGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}

	lines, totalHT, totalVAT, err := s.reportingAggregate(ctx, req.UserId, kind, period)
	if err != nil {
		return &invoiceGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}

	result, err := s.reporter.SubmitReport(ctx, pdp.SubmitReportInput{
		UserID:        req.UserId,
		Kind:          kind,
		Period:        period,
		Lines:         lines,
		TotalHTCents:  totalHT,
		TotalVATCents: totalVAT,
	})
	if err != nil {
		return &invoiceGrpc.GenericResponse{Success: false, Code: codes.ReportSubmissionFailed}, nil
	}
	target, ok := pdp.ToLifecycleStatus(result.Status)
	if !ok || target != "DEPOSITED" {
		return &invoiceGrpc.GenericResponse{Success: false, Code: codes.ReportSubmissionFailed}, nil
	}

	// Upsert the report row. ON CONFLICT keeps idempotency on re-submission of a
	// non-terminal period: a fresh aggregate replaces totals and resets status to
	// DEPOSITED (a new platform deposit), refreshing the platform handle.
	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO invoice_reports
		     (user_id, kind, period_year, period_month, status, total_ht_cents, total_vat_cents, report_id, submitted_at, updated_at)
		 VALUES ($1, $2, $3, $4, $5, $6, $7, NULLIF($8, ''), NOW(), NOW())
		 ON CONFLICT (user_id, kind, period_year, period_month) DO UPDATE
		     SET status          = EXCLUDED.status,
		         total_ht_cents  = EXCLUDED.total_ht_cents,
		         total_vat_cents = EXCLUDED.total_vat_cents,
		         report_id       = COALESCE(EXCLUDED.report_id, invoice_reports.report_id),
		         submitted_at    = NOW(),
		         updated_at      = NOW()`,
		req.UserId, string(kind), period.Year, period.Month, target,
		totalHT, totalVAT, strings.TrimSpace(result.ReportID),
	); err != nil {
		return &invoiceGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}

	return &invoiceGrpc.GenericResponse{Success: true, Code: codes.Success}, nil
}

// ListInvoiceReports returns the issuer's e-reports, most recent period first,
// optionally filtered by kind.
func (s *Server) ListInvoiceReports(ctx context.Context, req *invoiceGrpc.ListInvoiceReportsRequest) (resp *invoiceGrpc.ListInvoiceReportsResponse, err error) {
	startedAt := time.Now()
	defer deferObserve("list_invoice_reports", startedAt, func() (int32, bool) {
		if resp == nil {
			return codes.InternalError, false
		}
		return resp.Code, resp.Success
	}, &err)()

	if req == nil || strings.TrimSpace(req.UserId) == "" {
		return &invoiceGrpc.ListInvoiceReportsResponse{Success: false, Code: codes.InvalidInput}, nil
	}
	// Empty kind = no filter. An invalid non-empty kind yields an empty list.
	kindFilter := strings.TrimSpace(req.GetKind())

	rows, err := s.db.QueryContext(ctx,
		`SELECT kind, period_year, period_month, status, total_ht_cents, total_vat_cents, submitted_at
		   FROM invoice_reports
		  WHERE user_id=$1 AND ($2 = '' OR kind=$2)
		  ORDER BY period_year DESC, period_month DESC, kind`,
		req.UserId, kindFilter,
	)
	if err != nil {
		return &invoiceGrpc.ListInvoiceReportsResponse{Success: false, Code: codes.InternalError}, err
	}
	defer rows.Close()

	reports := make([]*invoiceGrpc.InvoiceReportSummary, 0)
	for rows.Next() {
		var kind, status string
		var year, month int32
		var totalHT, totalVAT int64
		var submittedAt time.Time
		if err := rows.Scan(&kind, &year, &month, &status, &totalHT, &totalVAT, &submittedAt); err != nil {
			return &invoiceGrpc.ListInvoiceReportsResponse{Success: false, Code: codes.InternalError}, err
		}
		reports = append(reports, &invoiceGrpc.InvoiceReportSummary{
			Kind:          kind,
			Year:          year,
			Month:         month,
			Status:        status,
			TotalHtCents:  totalHT,
			TotalVatCents: totalVAT,
			SubmittedAt:   submittedAt.Format(time.RFC3339),
		})
	}
	if err := rows.Err(); err != nil {
		return &invoiceGrpc.ListInvoiceReportsResponse{Success: false, Code: codes.InternalError}, err
	}
	return &invoiceGrpc.ListInvoiceReportsResponse{Success: true, Code: codes.Success, Reports: reports}, nil
}
