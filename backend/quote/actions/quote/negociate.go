package quote

import (
	"context"
	"database/sql"

	"project-devis-quote/actions/codes"
	quoteGrpc "project-devis-quote/services/grpc"
)

// Negociate moves a quote into the 'negociation' state. Allowed from 'draft':
// entering negotiation is the act of sending the quote to the client, so the
// gateway uses the returned client_id/name to email the client.
func Negociate(ctx context.Context, db *sql.DB, req *quoteGrpc.NegociateQuoteRequest) (*quoteGrpc.NegociateQuoteResponse, error) {
	if req.QuoteId == "" || req.UserId == "" {
		return &quoteGrpc.NegociateQuoteResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	var clientID, name string
	err := db.QueryRowContext(ctx,
		`UPDATE quotes SET state='negociation', updated_at=NOW()
		 WHERE quote_id=$1 AND user_id=$2 AND archived_at IS NULL AND state='draft'
		 RETURNING client_id, name`,
		req.QuoteId, req.UserId,
	).Scan(&clientID, &name)
	if err == sql.ErrNoRows {
		return &quoteGrpc.NegociateQuoteResponse{Success: false, Code: codes.NotFound}, nil
	}
	if err != nil {
		return &quoteGrpc.NegociateQuoteResponse{Success: false, Code: codes.InternalError}, err
	}

	return &quoteGrpc.NegociateQuoteResponse{
		Success:  true,
		Code:     codes.Success,
		ClientId: clientID,
		Name:     name,
	}, nil
}
