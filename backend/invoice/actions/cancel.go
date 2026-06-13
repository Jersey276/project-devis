package actions

import (
	"context"
	"strings"
	"time"

	"project-devis-invoice/actions/codes"
	invoiceGrpc "project-devis-invoice/services/grpc"
)

// CancelInvoice marks an ISSUED/PAID invoice as CANCELLED. The invoice and its
// number are preserved (no gap in the sequence). A DRAFT cannot be cancelled —
// it can simply be left or, later, deleted.
//
// NOTE (legal): strict FR compliance replaces a cancellation with a numbered
// credit note (avoir). CANCELLED is the MVP; see the credit-note follow-up.
func (s *Server) CancelInvoice(ctx context.Context, req *invoiceGrpc.CancelInvoiceRequest) (resp *invoiceGrpc.GenericResponse, err error) {
	startedAt := time.Now()
	defer deferObserve("cancel_invoice", startedAt, func() (int32, bool) {
		if resp == nil {
			return codes.InternalError, false
		}
		return resp.Code, resp.Success
	}, &err)()

	if req == nil || strings.TrimSpace(req.InvoiceId) == "" || strings.TrimSpace(req.UserId) == "" {
		return &invoiceGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	res, err := s.db.ExecContext(ctx,
		`UPDATE invoices SET status='CANCELLED', cancelled_at=NOW(), updated_at=NOW()
		 WHERE invoice_id=$1 AND user_id=$2 AND status IN ('ISSUED','PAID')`,
		req.InvoiceId, req.UserId,
	)
	if err != nil {
		return &invoiceGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	return s.replyOnAffected(ctx, req.InvoiceId, req.UserId, res)
}

// MarkInvoicePaid marks an ISSUED invoice as PAID.
func (s *Server) MarkInvoicePaid(ctx context.Context, req *invoiceGrpc.MarkInvoicePaidRequest) (resp *invoiceGrpc.GenericResponse, err error) {
	startedAt := time.Now()
	defer deferObserve("mark_invoice_paid", startedAt, func() (int32, bool) {
		if resp == nil {
			return codes.InternalError, false
		}
		return resp.Code, resp.Success
	}, &err)()

	if req == nil || strings.TrimSpace(req.InvoiceId) == "" || strings.TrimSpace(req.UserId) == "" {
		return &invoiceGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	res, err := s.db.ExecContext(ctx,
		`UPDATE invoices SET status='PAID', paid_at=NOW(), updated_at=NOW()
		 WHERE invoice_id=$1 AND user_id=$2 AND status='ISSUED'`,
		req.InvoiceId, req.UserId,
	)
	if err != nil {
		return &invoiceGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	return s.replyOnAffected(ctx, req.InvoiceId, req.UserId, res)
}
