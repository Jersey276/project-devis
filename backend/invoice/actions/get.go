package actions

import (
	"context"
	"database/sql"
	"sort"
	"strings"
	"time"

	"github.com/lib/pq"

	"project-devis-invoice/actions/codes"
	invoiceGrpc "project-devis-invoice/services/grpc"
)

func (s *Server) GetInvoice(ctx context.Context, req *invoiceGrpc.GetInvoiceRequest) (resp *invoiceGrpc.GetInvoiceResponse, err error) {
	startedAt := time.Now()
	defer deferObserve("get_invoice", startedAt, func() (int32, bool) {
		if resp == nil {
			return codes.InternalError, false
		}
		return resp.Code, resp.Success
	}, &err)()

	if req == nil || strings.TrimSpace(req.InvoiceId) == "" || strings.TrimSpace(req.UserId) == "" {
		return &invoiceGrpc.GetInvoiceResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	var (
		quoteID      string
		scheduleID   sql.NullString
		months       pq.Int32Array
		status       string
		invoiceNum   sql.NullString
		issuedAt     sql.NullTime
		saleDate     sql.NullTime
		dueDate      sql.NullTime
		totalHT      int64
		totalVAT     int64
		totalTTC     int64
		vatExempt    bool
	)
	err = s.db.QueryRowContext(ctx,
		`SELECT quote_id, schedule_id, billed_month_indexes, status, invoice_number,
		        issued_at, sale_date, due_date, total_ht_cents, total_vat_cents, total_ttc_cents, vat_exempt
		 FROM invoices WHERE invoice_id=$1 AND user_id=$2`,
		req.InvoiceId, req.UserId,
	).Scan(&quoteID, &scheduleID, &months, &status, &invoiceNum,
		&issuedAt, &saleDate, &dueDate, &totalHT, &totalVAT, &totalTTC, &vatExempt)
	if err == sql.ErrNoRows {
		return &invoiceGrpc.GetInvoiceResponse{Success: false, Code: codes.NotFound}, nil
	}
	if err != nil {
		return &invoiceGrpc.GetInvoiceResponse{Success: false, Code: codes.InternalError}, err
	}

	details := &invoiceGrpc.InvoiceDetails{
		InvoiceId:          req.InvoiceId,
		UserId:             req.UserId,
		QuoteId:            quoteID,
		ScheduleId:         scheduleID.String,
		BilledMonthIndexes: months,
		Status:             status,
		InvoiceNumber:      invoiceNum.String,
		IssuedAt:           formatNullTime(issuedAt, time.RFC3339),
		SaleDate:           formatNullTime(saleDate, "2006-01-02"),
		DueDate:            formatNullTime(dueDate, "2006-01-02"),
		TotalHtCents:       totalHT,
		TotalVatCents:      totalVAT,
		TotalTtcCents:      totalTTC,
		VatExempt:          vatExempt,
	}

	// DRAFT invoices have no frozen snapshot yet: return a live preview.
	if status == "DRAFT" {
		return s.getDraftPreview(ctx, req, details, quoteID, scheduleID, months)
	}

	// ISSUED/PAID/CANCELLED: read the frozen snapshot.
	if err := s.loadSnapshot(ctx, req.InvoiceId, details); err != nil {
		return &invoiceGrpc.GetInvoiceResponse{Success: false, Code: codes.InternalError}, err
	}

	// Expose which lines are already credited so the UI can grey them out.
	if status == "ISSUED" || status == "PAID" {
		credited, err := s.creditedPositions(ctx, req.InvoiceId)
		if err != nil {
			return &invoiceGrpc.GetInvoiceResponse{Success: false, Code: codes.InternalError}, err
		}
		positions := make([]int32, 0, len(credited))
		for p := range credited {
			positions = append(positions, p)
		}
		sort.Slice(positions, func(i, j int) bool { return positions[i] < positions[j] })
		details.CreditedPositions = positions
	}
	return &invoiceGrpc.GetInvoiceResponse{Success: true, Code: codes.Success, Invoice: details}, nil
}

// getDraftPreview resolves the source live (the quote/schedule is frozen by its
// own validation) so the user can review a draft before issuing.
func (s *Server) getDraftPreview(ctx context.Context, req *invoiceGrpc.GetInvoiceRequest, details *invoiceGrpc.InvoiceDetails, quoteID string, scheduleID sql.NullString, months pq.Int32Array) (*invoiceGrpc.GetInvoiceResponse, error) {
	var (
		resolved *resolvedInvoice
		code     int32
		err      error
	)
	if scheduleID.Valid && scheduleID.String != "" {
		resolved, code, err = s.resolveScheduleInvoice(ctx, req.UserId, quoteID, scheduleID.String, months)
	} else {
		resolved, code, err = s.resolveQuoteInvoice(ctx, req.UserId, quoteID)
	}
	if err != nil {
		return &invoiceGrpc.GetInvoiceResponse{Success: false, Code: code}, err
	}
	if code != codes.Success {
		return &invoiceGrpc.GetInvoiceResponse{Success: false, Code: code}, nil
	}

	totals := computeTotals(resolved.compute, resolved.vatExempt)
	details.Issuer = partyToProto(resolved.parties, true)
	details.Client = partyToProto(resolved.parties, false)
	details.VatExempt = resolved.vatExempt
	details.TotalHtCents = totals.totalHT
	details.TotalVatCents = totals.totalVAT
	details.TotalTtcCents = totals.totalTTC
	for _, l := range resolved.lines {
		details.Lines = append(details.Lines, &invoiceGrpc.InvoiceLine{
			QuoteLineId:    l.quoteLineID,
			Name:           l.name,
			Unit:           l.unit,
			Quantity:       l.quantity,
			UnitPriceCents: l.unitPriceCents,
			LineHtCents:    l.lineHTCents,
			TaxId:          l.taxID,
			TaxRate:        l.taxRate,
			TaxLabel:       l.taxLabel,
		})
	}
	for _, b := range totals.breakdown {
		details.VatBreakdown = append(details.VatBreakdown, &invoiceGrpc.InvoiceVatLine{
			TaxRate: b.rate, BaseHtCents: b.baseHT, VatCents: b.vat,
		})
	}
	return &invoiceGrpc.GetInvoiceResponse{Success: true, Code: codes.Success, Invoice: details}, nil
}

func formatNullTime(t sql.NullTime, layout string) string {
	if !t.Valid {
		return ""
	}
	return t.Time.Format(layout)
}
