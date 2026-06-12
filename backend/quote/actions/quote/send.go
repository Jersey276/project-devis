package quote

import (
	"context"
	"database/sql"

	"project-devis-quote/actions/codes"
	quoteGrpc "project-devis-quote/services/grpc"
)

func Send(ctx context.Context, db *sql.DB, req *quoteGrpc.SendQuoteRequest) (*quoteGrpc.SendQuoteResponse, error) {
	if req.QuoteId == "" || req.UserId == "" {
		return &quoteGrpc.SendQuoteResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	var clientID, name string
	err := db.QueryRowContext(ctx,
		`UPDATE quotes SET state='sent', updated_at=NOW()
		 WHERE quote_id=$1 AND user_id=$2 AND archived_at IS NULL AND state='draft'
		 RETURNING client_id, name`,
		req.QuoteId, req.UserId,
	).Scan(&clientID, &name)
	if err == sql.ErrNoRows {
		return &quoteGrpc.SendQuoteResponse{Success: false, Code: codes.NotFound}, nil
	}
	if err != nil {
		return &quoteGrpc.SendQuoteResponse{Success: false, Code: codes.InternalError}, err
	}

	return &quoteGrpc.SendQuoteResponse{
		Success:  true,
		Code:     codes.Success,
		ClientId: clientID,
		Name:     name,
	}, nil
}
