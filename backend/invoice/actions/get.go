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
		quoteID    string
		scheduleID sql.NullString
		months     pq.Int32Array
		status     string
		invoiceNum sql.NullString
		issuedAt   sql.NullTime
		saleDate   sql.NullTime
		dueDate    sql.NullTime
		totalHT    int64
		totalVAT   int64
		totalTTC   int64
		vatExempt  bool
		lifecycle  string
	)
	var query string
	var queryArgs []interface{}
	if strings.TrimSpace(req.ClientId) != "" {
		// Customer-facing read: verify the invoice belongs to the given client via
		// its quote (invoices don't store client_id directly).
		query = `SELECT i.quote_id, i.schedule_id, i.billed_month_indexes, i.status,
		                i.invoice_number, i.issued_at, i.sale_date, i.due_date,
		                i.total_ht_cents, i.total_vat_cents, i.total_ttc_cents,
		                i.vat_exempt, i.lifecycle_status
		         FROM invoices i
		         JOIN quotes q ON q.quote_id = i.quote_id
		         WHERE i.invoice_id=$1 AND i.user_id=$2 AND q.client_id=$3`
		queryArgs = []interface{}{req.InvoiceId, req.UserId, req.ClientId}
	} else {
		query = `SELECT quote_id, schedule_id, billed_month_indexes, status, invoice_number,
		                issued_at, sale_date, due_date, total_ht_cents, total_vat_cents, total_ttc_cents,
		                vat_exempt, lifecycle_status
		         FROM invoices WHERE invoice_id=$1 AND user_id=$2`
		queryArgs = []interface{}{req.InvoiceId, req.UserId}
	}

	err = s.db.QueryRowContext(ctx, query, queryArgs...).Scan(
		&quoteID, &scheduleID, &months, &status, &invoiceNum,
		&issuedAt, &saleDate, &dueDate, &totalHT, &totalVAT, &totalTTC, &vatExempt, &lifecycle)
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
		LifecycleStatus:    lifecycle,
	}

	if status == "DRAFT" {
		return s.getDraftPreview(ctx, req, details, quoteID, scheduleID, months)
	}

	if err := s.loadSnapshot(ctx, req.InvoiceId, details); err != nil {
		return &invoiceGrpc.GetInvoiceResponse{Success: false, Code: codes.InternalError}, err
	}

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

func (s *Server) getDraftPreview(ctx context.Context, req *invoiceGrpc.GetInvoiceRequest, details *invoiceGrpc.InvoiceDetails, quoteID string, scheduleID sql.NullString, months pq.Int32Array) (*invoiceGrpc.GetInvoiceResponse, error) {
	var (
		resolved *resolvedInvoice
		code     int32
		err      error
	)

	previewAt := time.Now().In(invoiceTZ)
	if scheduleID.Valid && scheduleID.String != "" {
		resolved, code, err = s.resolveScheduleInvoice(ctx, req.InvoiceId, req.UserId, quoteID, scheduleID.String, months, previewAt)
	} else {
		resolved, code, err = s.resolveQuoteInvoice(ctx, req.InvoiceId, req.UserId, quoteID, previewAt)
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
