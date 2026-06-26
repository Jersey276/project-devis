package quote

import (
	"context"
	"database/sql"

	"project-devis-quote/actions/codes"
	quoteGrpc "project-devis-quote/services/grpc"
)

// Validate moves a quote to 'validated'. Allowed from 'negociation'.
// 'validated' is terminal: it locks editing and unlocks invoicing.
func Validate(ctx context.Context, db *sql.DB, req *quoteGrpc.ValidateQuoteRequest) (*quoteGrpc.GenericResponse, error) {
	if req.QuoteId == "" || req.UserId == "" {
		return &quoteGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	res, err := db.ExecContext(ctx,
		`UPDATE quotes SET state='validated', updated_at=NOW()
		 WHERE quote_id=$1 AND user_id=$2 AND archived_at IS NULL
		   AND state='negociation'`,
		req.QuoteId, req.UserId,
	)
	if err != nil {
		return &quoteGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &quoteGrpc.GenericResponse{Success: false, Code: codes.NotFound}, nil
	}

	return &quoteGrpc.GenericResponse{Success: true, Code: codes.Success}, nil
}
