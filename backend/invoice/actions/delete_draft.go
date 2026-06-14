package actions

import (
	"context"
	"strings"
	"time"

	"project-devis-invoice/actions/codes"
	invoiceGrpc "project-devis-invoice/services/grpc"
)

// DeleteDraftInvoice removes an invoice that is still a DRAFT. A draft carries no
// legal number, no frozen snapshot and no seal, so deleting it is safe and leaves
// the inalterability chain untouched. The DELETE is guarded by status='DRAFT', so
// an already-issued (sealed) invoice is never affected — it stays immutable.
func (s *Server) DeleteDraftInvoice(ctx context.Context, req *invoiceGrpc.DeleteDraftInvoiceRequest) (resp *invoiceGrpc.GenericResponse, err error) {
	startedAt := time.Now()
	defer deferObserve("delete_draft_invoice", startedAt, func() (int32, bool) {
		if resp == nil {
			return codes.InternalError, false
		}
		return resp.Code, resp.Success
	}, &err)()

	if req == nil || strings.TrimSpace(req.InvoiceId) == "" || strings.TrimSpace(req.UserId) == "" {
		return &invoiceGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	res, err := s.db.ExecContext(ctx,
		`DELETE FROM invoices WHERE invoice_id=$1 AND user_id=$2 AND status='DRAFT'`,
		req.InvoiceId, req.UserId,
	)
	if err != nil {
		return &invoiceGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	// 0 rows: either the invoice does not exist (NotFound) or it exists but is no
	// longer DRAFT (InvoiceFinalized — issued, cannot be deleted).
	return s.replyOnAffected(ctx, req.InvoiceId, req.UserId, res)
}
