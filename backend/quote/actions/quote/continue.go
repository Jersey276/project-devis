package quote

import (
	"context"
	"database/sql"

	"project-devis-quote/actions/codes"
	quoteGrpc "project-devis-quote/services/grpc"
)

func Continue(ctx context.Context, db *sql.DB, req *quoteGrpc.ContinueQuoteRequest) (*quoteGrpc.GenericResponse, error) {
	if req.QuoteId == "" || req.UserId == "" {
		return &quoteGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	res, err := db.ExecContext(ctx,
		`UPDATE quotes SET state='draft', updated_at=NOW()
		 WHERE quote_id=$1 AND user_id=$2 AND state='drop'`,
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
