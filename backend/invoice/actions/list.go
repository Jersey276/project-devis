package actions

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"project-devis-invoice/actions/codes"
	invoiceGrpc "project-devis-invoice/services/grpc"
)

func (s *Server) ListInvoices(ctx context.Context, req *invoiceGrpc.ListInvoicesRequest) (resp *invoiceGrpc.ListInvoicesResponse, err error) {
	startedAt := time.Now()
	defer deferObserve("list_invoices", startedAt, func() (int32, bool) {
		if resp == nil {
			return codes.InternalError, false
		}
		return resp.Code, resp.Success
	}, &err)()

	if req == nil || strings.TrimSpace(req.UserId) == "" {
		return &invoiceGrpc.ListInvoicesResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	query := `SELECT invoice_id, invoice_number, status, quote_id, schedule_id,
	                 issued_at, due_date, total_ttc_cents
	          FROM invoices WHERE user_id=$1`
	args := []any{req.UserId}
	if q := strings.TrimSpace(req.QuoteId); q != "" {
		query += ` AND quote_id=$2`
		args = append(args, q)
	}
	query += ` ORDER BY created_at DESC`

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return &invoiceGrpc.ListInvoicesResponse{Success: false, Code: codes.InternalError}, err
	}
	defer rows.Close()

	out := make([]*invoiceGrpc.InvoiceSummary, 0)
	for rows.Next() {
		var (
			id, status, quoteID string
			number              sql.NullString
			scheduleID          sql.NullString
			issuedAt, dueDate   sql.NullTime
			totalTTC            int64
		)
		if err := rows.Scan(&id, &number, &status, &quoteID, &scheduleID, &issuedAt, &dueDate, &totalTTC); err != nil {
			return &invoiceGrpc.ListInvoicesResponse{Success: false, Code: codes.InternalError}, err
		}
		out = append(out, &invoiceGrpc.InvoiceSummary{
			InvoiceId:     id,
			InvoiceNumber: number.String,
			Status:        status,
			QuoteId:       quoteID,
			ScheduleId:    scheduleID.String,
			IssuedAt:      formatNullTime(issuedAt, time.RFC3339),
			DueDate:       formatNullTime(dueDate, "2006-01-02"),
			TotalTtcCents: totalTTC,
		})
	}
	if err := rows.Err(); err != nil {
		return &invoiceGrpc.ListInvoicesResponse{Success: false, Code: codes.InternalError}, err
	}

	return &invoiceGrpc.ListInvoicesResponse{Success: true, Code: codes.Success, Invoices: out}, nil
}
