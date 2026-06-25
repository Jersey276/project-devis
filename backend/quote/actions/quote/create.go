package quote

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"project-devis-quote/actions/codes"
	"project-devis-quote/actions/sqlutil"
	quoteGrpc "project-devis-quote/services/grpc"
)

func Create(ctx context.Context, db *sql.DB, req *quoteGrpc.CreateQuoteRequest) (*quoteGrpc.CreateQuoteResponse, error) {
	var fieldErrors []*quoteGrpc.ValidationError

	if req.UserId == "" {
		fieldErrors = append(fieldErrors, sqlutil.Required("user_id"))
	}
	if req.Name == "" {
		fieldErrors = append(fieldErrors, sqlutil.Required("name"))
	}
	if req.ClientId == "" {
		fieldErrors = append(fieldErrors, sqlutil.Required("client_id"))
	}
	if req.AddressId == 0 {
		fieldErrors = append(fieldErrors, sqlutil.Required("address_id"))
	}
	if req.UserAddressId == 0 {
		fieldErrors = append(fieldErrors, sqlutil.Required("user_address_id"))
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
