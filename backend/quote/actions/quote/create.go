package quote

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"project-devis-quote/actions/codes"
	quoteGrpc "project-devis-quote/services/grpc"
)

func Create(ctx context.Context, db *sql.DB, req *quoteGrpc.CreateQuoteRequest) (*quoteGrpc.CreateQuoteResponse, error) {
	if req.UserId == "" || req.Name == "" || req.ClientId == "" || req.AddressId == 0 {
		return &quoteGrpc.CreateQuoteResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	quoteID := uuid.New().String()
	_, err := db.ExecContext(ctx,
		`INSERT INTO quotes (quote_id, user_id, name, state, client_id, address_id) VALUES ($1, $2, $3, 'draft', $4, $5)`,
		quoteID, req.UserId, req.Name, req.ClientId, req.AddressId,
	)
	if err != nil {
		return &quoteGrpc.CreateQuoteResponse{Success: false, Code: codes.InternalError}, err
	}

	return &quoteGrpc.CreateQuoteResponse{Success: true, Code: codes.Success, QuoteId: quoteID}, nil
}
