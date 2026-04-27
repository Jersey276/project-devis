package quote

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"project-devis-quote/actions/codes"
	quoteGrpc "project-devis-quote/services/grpc"
)

func Create(ctx context.Context, db *sql.DB, req *quoteGrpc.CreateQuoteRequest) (*quoteGrpc.CreateQuoteResponse, error) {
	if req.UserId == "" || req.Name == "" {
		return &quoteGrpc.CreateQuoteResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	quoteID := uuid.New().String()
	_, err := db.ExecContext(ctx,
		`INSERT INTO quotes (quote_id, user_id, name, state) VALUES ($1, $2, $3, 'draft')`,
		quoteID, req.UserId, req.Name,
	)
	if err != nil {
		return &quoteGrpc.CreateQuoteResponse{Success: false, Code: codes.InternalError}, err
	}

	return &quoteGrpc.CreateQuoteResponse{Success: true, Code: codes.Success, QuoteId: quoteID}, nil
}
