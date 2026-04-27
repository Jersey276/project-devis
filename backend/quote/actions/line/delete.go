package line

import (
	"context"
	"database/sql"

	"project-devis-quote/actions/codes"
	quoteGrpc "project-devis-quote/services/grpc"
)

func Delete(ctx context.Context, db *sql.DB, req *quoteGrpc.DeleteQuoteLineRequest) (*quoteGrpc.GenericResponse, error) {
	if req.LineId == "" || req.UserId == "" {
		return &quoteGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	res, err := db.ExecContext(ctx,
		`DELETE FROM quote_lines l
		 USING quotes q
		 WHERE l.quote_id = q.quote_id
		   AND l.line_id = $1
		   AND q.user_id = $2`,
		req.LineId, req.UserId,
	)
	if err != nil {
		return &quoteGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &quoteGrpc.GenericResponse{Success: false, Code: codes.NotFound}, nil
	}

	return &quoteGrpc.GenericResponse{Success: true, Code: codes.Success}, nil
}
