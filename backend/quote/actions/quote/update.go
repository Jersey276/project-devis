package quote

import (
	"context"
	"database/sql"

	"project-devis-quote/actions/codes"
	quoteGrpc "project-devis-quote/services/grpc"
)

func Update(ctx context.Context, db *sql.DB, req *quoteGrpc.UpdateQuoteRequest) (*quoteGrpc.UpdateQuoteResponse, error) {
	if req.QuoteId == "" || req.UserId == "" || req.Name == "" {
		return &quoteGrpc.UpdateQuoteResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	if code, ok := EditableForUser(ctx, db, req.QuoteId, req.UserId); !ok {
		return &quoteGrpc.UpdateQuoteResponse{Success: false, Code: code}, nil
	}

	_, err := db.ExecContext(ctx,
		`UPDATE quotes SET name=$1, updated_at=NOW() WHERE quote_id=$2 AND user_id=$3`,
		req.Name, req.QuoteId, req.UserId,
	)
	if err != nil {
		return &quoteGrpc.UpdateQuoteResponse{Success: false, Code: codes.InternalError}, err
	}

	return &quoteGrpc.UpdateQuoteResponse{Success: true, Code: codes.Success}, nil
}
