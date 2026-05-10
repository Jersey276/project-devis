package quote

import (
	"context"
	"database/sql"

	"project-devis-quote/actions/codes"
	quoteGrpc "project-devis-quote/services/grpc"
)

// Update accepts a Name (required) plus optional ClientId / AddressId. An
// empty ClientId or zero AddressId is the "preserve existing value" sentinel,
// so the caller can ship just `{ name }` without nulling the foreign keys.
// COALESCE(NULLIF(...)) keeps the SQL static — adding another optional column
// later is just one more line, no placeholder threading.
func Update(ctx context.Context, db *sql.DB, req *quoteGrpc.UpdateQuoteRequest) (*quoteGrpc.UpdateQuoteResponse, error) {
	if req.QuoteId == "" || req.UserId == "" || req.Name == "" {
		return &quoteGrpc.UpdateQuoteResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	if code, ok := EditableForUser(ctx, db, req.QuoteId, req.UserId); !ok {
		return &quoteGrpc.UpdateQuoteResponse{Success: false, Code: code}, nil
	}

	_, err := db.ExecContext(ctx,
		`UPDATE quotes SET
		    name       = $3,
		    client_id  = COALESCE(NULLIF($4, '')::TEXT, client_id),
		    address_id = COALESCE(NULLIF($5, 0)::INT,  address_id),
		    updated_at = NOW()
		 WHERE quote_id = $1 AND user_id = $2`,
		req.QuoteId, req.UserId, req.Name, req.ClientId, req.AddressId,
	)
	if err != nil {
		return &quoteGrpc.UpdateQuoteResponse{Success: false, Code: codes.InternalError}, err
	}

	return &quoteGrpc.UpdateQuoteResponse{Success: true, Code: codes.Success}, nil
}
