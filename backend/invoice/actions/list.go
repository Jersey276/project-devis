package actions

import (
	"context"
	"database/sql"
	"fmt"
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

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	where, args := buildInvoiceFilters(req.UserId, req.QuoteId, req.Filters)

	var total int64
	if err = s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM invoices"+where, args...).Scan(&total); err != nil {
		return &invoiceGrpc.ListInvoicesResponse{Success: false, Code: codes.InternalError}, err
	}

	orderBy := buildInvoiceOrderBy(req.SortBy, req.SortDirection)

	args = append(args, pageSize, offset)
	n := len(args)
	query := fmt.Sprintf(
		`SELECT invoice_id, invoice_number, status, quote_id, schedule_id,
		        issued_at, due_date, total_ttc_cents, lifecycle_status
		 FROM invoices%s ORDER BY %s LIMIT $%d OFFSET $%d`,
		where, orderBy, n-1, n,
	)

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
			lifecycle           string
		)
		if err := rows.Scan(&id, &number, &status, &quoteID, &scheduleID, &issuedAt, &dueDate, &totalTTC, &lifecycle); err != nil {
			return &invoiceGrpc.ListInvoicesResponse{Success: false, Code: codes.InternalError}, err
		}
		out = append(out, &invoiceGrpc.InvoiceSummary{
			InvoiceId:       id,
			InvoiceNumber:   number.String,
			Status:          status,
			QuoteId:         quoteID,
			ScheduleId:      scheduleID.String,
			IssuedAt:        formatNullTime(issuedAt, time.RFC3339),
			DueDate:         formatNullTime(dueDate, "2006-01-02"),
			TotalTtcCents:   totalTTC,
			LifecycleStatus: lifecycle,
		})
	}
	if err := rows.Err(); err != nil {
		return &invoiceGrpc.ListInvoicesResponse{Success: false, Code: codes.InternalError}, err
	}

	return &invoiceGrpc.ListInvoicesResponse{Success: true, Code: codes.Success, Invoices: out, Total: total}, nil
}

var allowedInvoiceSortColumns = map[string]string{
	"number":    "invoice_number",
	"status":    "status",
	"lifecycle": "lifecycle_status",
	"quoteId":   "quote_id",
	"dueDate":   "due_date",
}

func buildInvoiceOrderBy(sortBy, sortDirection string) string {
	return buildOrderBy(allowedInvoiceSortColumns, "created_at", sortBy, sortDirection)
}

// buildOrderBy maps a frontend column name to its DB equivalent via the
// allowed whitelist and appends the sanitised direction. Falls back to
// defaultCol when sortBy is not in the whitelist.
func buildOrderBy(allowed map[string]string, defaultCol, sortBy, sortDirection string) string {
	col, ok := allowed[sortBy]
	if !ok {
		col = defaultCol
	}
	if strings.ToUpper(sortDirection) == "ASC" {
		return col + " ASC"
	}
	return col + " DESC"
}

func buildInvoiceFilters(userID, legacyQuoteID string, f *invoiceGrpc.InvoiceFilters) (string, []interface{}) {
	args := []interface{}{userID}
	clauses := []string{"user_id = $1"}

	if strings.TrimSpace(legacyQuoteID) != "" {
		args = append(args, legacyQuoteID)
		clauses = append(clauses, fmt.Sprintf("quote_id = $%d", len(args)))
	}

	if f == nil {
		return " WHERE " + strings.Join(clauses, " AND "), args
	}

	if len(f.Statuses) > 0 {
		placeholders := make([]string, len(f.Statuses))
		for i, s := range f.Statuses {
			args = append(args, s)
			placeholders[i] = fmt.Sprintf("$%d", len(args))
		}
		clauses = append(clauses, "status IN ("+strings.Join(placeholders, ",")+")")
	}
	if len(f.LifecycleStatuses) > 0 {
		placeholders := make([]string, len(f.LifecycleStatuses))
		for i, s := range f.LifecycleStatuses {
			args = append(args, s)
			placeholders[i] = fmt.Sprintf("$%d", len(args))
		}
		clauses = append(clauses, "lifecycle_status IN ("+strings.Join(placeholders, ",")+")")
	}
	if f.IssuedFrom != "" {
		args = append(args, f.IssuedFrom)
		clauses = append(clauses, fmt.Sprintf("issued_at >= $%d", len(args)))
	}
	if f.IssuedTo != "" {
		args = append(args, f.IssuedTo)
		clauses = append(clauses, fmt.Sprintf("issued_at <= $%d", len(args)))
	}
	if f.DueFrom != "" {
		args = append(args, f.DueFrom)
		clauses = append(clauses, fmt.Sprintf("due_date >= $%d", len(args)))
	}
	if f.DueTo != "" {
		args = append(args, f.DueTo)
		clauses = append(clauses, fmt.Sprintf("due_date <= $%d", len(args)))
	}
	if f.QuoteIdFilter != "" {
		args = append(args, f.QuoteIdFilter)
		clauses = append(clauses, fmt.Sprintf("quote_id = $%d", len(args)))
	}
	// client_id filter via subquery (invoices don't store client_id directly)
	if f.ClientId != "" {
		args = append(args, f.ClientId)
		clauses = append(clauses, fmt.Sprintf(
			"quote_id IN (SELECT quote_id FROM quotes WHERE client_id = $%d)", len(args),
		))
	}

	return " WHERE " + strings.Join(clauses, " AND "), args
}
