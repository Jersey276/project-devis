package quote

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"project-devis-quote/actions/codes"
	quoteGrpc "project-devis-quote/services/grpc"
)

func Create(ctx context.Context, db *sql.DB, req *quoteGrpc.CreateQuoteRequest) (*quoteGrpc.CreateQuoteResponse, error) {
	var fieldErrors []*quoteGrpc.ValidationError

	if req.UserId == "" {
		fieldErrors = append(fieldErrors, &quoteGrpc.ValidationError{Field: "user_id", Message: "Champ requis."})
	}
	if req.Name == "" {
		fieldErrors = append(fieldErrors, &quoteGrpc.ValidationError{Field: "name", Message: "Champ requis."})
	}
	if req.ClientId == "" {
		fieldErrors = append(fieldErrors, &quoteGrpc.ValidationError{Field: "client_id", Message: "Champ requis."})
	}
	if req.AddressId == 0 {
		fieldErrors = append(fieldErrors, &quoteGrpc.ValidationError{Field: "address_id", Message: "Champ requis."})
	}
	if req.UserAddressId == 0 {
		fieldErrors = append(fieldErrors, &quoteGrpc.ValidationError{Field: "user_address_id", Message: "Champ requis."})
	}

	if len(fieldErrors) > 0 {
		return &quoteGrpc.CreateQuoteResponse{Success: false, Code: codes.InvalidInput, ValidationErrors: fieldErrors}, nil
	}

	quoteID := uuid.New().String()
	_, err := db.ExecContext(ctx,
		`INSERT INTO quotes (quote_id, user_id, name, state, client_id, address_id, user_address_id) VALUES ($1, $2, $3, 'draft', $4, $5, $6)`,
		quoteID, req.UserId, req.Name, req.ClientId, req.AddressId, req.UserAddressId,
	)
	if err != nil {
		return &quoteGrpc.CreateQuoteResponse{Success: false, Code: codes.InternalError}, err
	}

	return &quoteGrpc.CreateQuoteResponse{Success: true, Code: codes.Success, QuoteId: quoteID}, nil
}
