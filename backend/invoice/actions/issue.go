package actions

import (
	"context"
	"database/sql"
	"strings"
	"time"

	"github.com/lib/pq"

	"project-devis-invoice/actions/codes"
	invoiceGrpc "project-devis-invoice/services/grpc"
)

// IssueInvoice promotes a DRAFT invoice to ISSUED: it assigns the legal number,
// freezes the snapshot and locks the figures. Idempotent if already issued.
func (s *Server) IssueInvoice(ctx context.Context, req *invoiceGrpc.IssueInvoiceRequest) (resp *invoiceGrpc.CreateInvoiceResponse, err error) {
	startedAt := time.Now()
	defer deferObserve("issue_invoice", startedAt, func() (int32, bool) {
		if resp == nil {
			return codes.InternalError, false
		}
		return resp.Code, resp.Success
	}, &err)()

	if req == nil || strings.TrimSpace(req.InvoiceId) == "" || strings.TrimSpace(req.UserId) == "" {
		return &invoiceGrpc.CreateInvoiceResponse{Success: false, Code: codes.InvalidInput}, nil
	}
	return s.issue(ctx, req.InvoiceId, req.UserId, "", 0)
}

// issue performs the DRAFT → ISSUED transition. saleDate/dueInDays are taken
// from the create request when issuing in one step; on a standalone issue they
// are empty/zero and defaults apply.
func (s *Server) issue(ctx context.Context, invoiceID, userID, saleDate string, dueInDays int32) (*invoiceGrpc.CreateInvoiceResponse, error) {
	// Load the current row to learn the source and guard the state.
	var (
		quoteID      string
		scheduleID   sql.NullString
		status       string
		months       pq.Int32Array
		existingNum  sql.NullString
	)
	err := s.db.QueryRowContext(ctx,
		`SELECT quote_id, schedule_id, status, billed_month_indexes, invoice_number
		 FROM invoices WHERE invoice_id=$1 AND user_id=$2`,
		invoiceID, userID,
	).Scan(&quoteID, &scheduleID, &status, &months, &existingNum)
	if err == sql.ErrNoRows {
		return &invoiceGrpc.CreateInvoiceResponse{Success: false, Code: codes.NotFound}, nil
	}
	if err != nil {
		return &invoiceGrpc.CreateInvoiceResponse{Success: false, Code: codes.InternalError}, err
	}

	switch status {
	case "ISSUED", "PAID":
		// Idempotent: already issued, return the existing number.
		return &invoiceGrpc.CreateInvoiceResponse{Success: true, Code: codes.Success, InvoiceId: invoiceID, InvoiceNumber: existingNum.String}, nil
	case "CANCELLED":
		return &invoiceGrpc.CreateInvoiceResponse{Success: false, Code: codes.InvoiceFinalized}, nil
	}

	// Fix the legal issue date up front: it scopes the OSS threshold cumulative
	// (Europe/Paris civil year) computed during resolution as well as the
	// numbering year below.
	issuedAt := time.Now().In(invoiceTZ)

	// Resolve the billable data from the source (cells for schedule, full lines
	// for quote).
	var (
		resolved *resolvedInvoice
		code     int32
	)
	if scheduleID.Valid && scheduleID.String != "" {
		resolved, code, err = s.resolveScheduleInvoice(ctx, invoiceID, userID, quoteID, scheduleID.String, months, issuedAt)
	} else {
		resolved, code, err = s.resolveQuoteInvoice(ctx, invoiceID, userID, quoteID, issuedAt)
	}
	if err != nil {
		return &invoiceGrpc.CreateInvoiceResponse{Success: false, Code: code}, err
	}
	if code != codes.Success {
		return &invoiceGrpc.CreateInvoiceResponse{Success: false, Code: code}, nil
	}
	if len(resolved.compute) == 0 {
		return &invoiceGrpc.CreateInvoiceResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	totals := computeTotals(resolved.compute, resolved.vatExempt)

	sale, due := resolveSaleAndDue(issuedAt, saleDate, dueInDays)
	year := issuedAt.Year()

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return &invoiceGrpc.CreateInvoiceResponse{Success: false, Code: codes.InternalError}, err
	}
	defer tx.Rollback()

	// Re-read the status under a row lock to defend against a concurrent issue.
	var lockedStatus string
	if err := tx.QueryRowContext(ctx,
		`SELECT status FROM invoices WHERE invoice_id=$1 FOR UPDATE`, invoiceID,
	).Scan(&lockedStatus); err != nil {
		return &invoiceGrpc.CreateInvoiceResponse{Success: false, Code: codes.InternalError}, err
	}
	if lockedStatus != "DRAFT" {
		// Someone issued it between our read and the lock — treat as idempotent.
		return &invoiceGrpc.CreateInvoiceResponse{Success: false, Code: codes.InvoiceFinalized}, nil
	}

	number, seq, err := allocateInvoiceNumber(ctx, tx, userID, year)
	if err != nil {
		return &invoiceGrpc.CreateInvoiceResponse{Success: false, Code: codes.InternalError}, err
	}

	if _, err := tx.ExecContext(ctx,
		`UPDATE invoices SET
			status='ISSUED', invoice_number=$2, number_year=$3, number_seq=$4,
			issued_at=$5, sale_date=$6, due_date=$7,
			total_ht_cents=$8, total_vat_cents=$9, total_ttc_cents=$10, vat_exempt=$11,
			updated_at=NOW()
		 WHERE invoice_id=$1`,
		invoiceID, number, year, seq,
		issuedAt, sale, due,
		totals.totalHT, totals.totalVAT, totals.totalTTC, resolved.vatExempt,
	); err != nil {
		return &invoiceGrpc.CreateInvoiceResponse{Success: false, Code: codes.InternalError}, err
	}

	if err := writeSnapshots(ctx, tx, invoiceID, resolved, totals.breakdown); err != nil {
		return &invoiceGrpc.CreateInvoiceResponse{Success: false, Code: codes.InternalError}, err
	}

	// Seal the invoice into the issuer's cryptographic chain (inalterability).
	contentHash := computeContentHash(sealableDoc{
		userID:    userID,
		docType:   "INVOICE",
		number:    number,
		issuedAt:  issuedAt,
		totalHT:   totals.totalHT,
		totalVAT:  totals.totalVAT,
		totalTTC:  totals.totalTTC,
		vatExempt: resolved.vatExempt,
		lines:     sealLinesFromSnapshots(resolved.lines),
	})
	if _, _, err := sealDocument(ctx, tx, userID, "INVOICE", invoiceID, contentHash); err != nil {
		return &invoiceGrpc.CreateInvoiceResponse{Success: false, Code: codes.SealError}, err
	}

	if err := tx.Commit(); err != nil {
		return &invoiceGrpc.CreateInvoiceResponse{Success: false, Code: codes.InternalError}, err
	}

	return &invoiceGrpc.CreateInvoiceResponse{Success: true, Code: codes.Success, InvoiceId: invoiceID, InvoiceNumber: number}, nil
}
