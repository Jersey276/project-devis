package actions

import (
	"context"
	"database/sql"

	"project-devis-invoice/actions/codes"
	invoiceGrpc "project-devis-invoice/services/grpc"
)

// replyOnAffected builds the response for a conditional status UPDATE: if a row
// changed it succeeded; otherwise we disambiguate between a missing invoice
// (NotFound) and one in a state the transition does not allow (InvoiceFinalized).
func (s *Server) replyOnAffected(ctx context.Context, invoiceID, userID string, res sql.Result) (*invoiceGrpc.GenericResponse, error) {
	affected, err := res.RowsAffected()
	if err != nil {
		return &invoiceGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	if affected > 0 {
		return &invoiceGrpc.GenericResponse{Success: true, Code: codes.Success}, nil
	}

	var exists bool
	if err := s.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM invoices WHERE invoice_id=$1 AND user_id=$2)`,
		invoiceID, userID,
	).Scan(&exists); err != nil {
		return &invoiceGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	if !exists {
		return &invoiceGrpc.GenericResponse{Success: false, Code: codes.NotFound}, nil
	}
	return &invoiceGrpc.GenericResponse{Success: false, Code: codes.InvoiceFinalized}, nil
}
