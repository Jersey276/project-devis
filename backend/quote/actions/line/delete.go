package line

import (
	"context"
	"database/sql"

	"project-devis-quote/actions/codes"
	"project-devis-quote/actions/quote"
	quoteGrpc "project-devis-quote/services/grpc"
)

func Delete(ctx context.Context, db *sql.DB, req *quoteGrpc.DeleteQuoteLineRequest) (*quoteGrpc.GenericResponse, error) {
	if req.LineId == "" || req.UserId == "" {
		return &quoteGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	if code, ok := quote.LineParentEditable(ctx, db, req.LineId, req.UserId); !ok {
		return &quoteGrpc.GenericResponse{Success: false, Code: code}, nil
	}

	res, err := db.ExecContext(ctx,
		`DELETE FROM quote_lines WHERE line_id=$1`,
		req.LineId,
	)
	if err != nil {
		return &quoteGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &quoteGrpc.GenericResponse{Success: false, Code: codes.NotFound}, nil
	}

	return &quoteGrpc.GenericResponse{Success: true, Code: codes.Success}, nil
}
