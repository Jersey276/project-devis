package quote

import (
	"context"
	"database/sql"

	"project-devis-quote/actions/codes"
	"project-devis-quote/actions/sqlutil"
	quoteGrpc "project-devis-quote/services/grpc"
)

func Update(ctx context.Context, db *sql.DB, req *quoteGrpc.UpdateQuoteRequest) (*quoteGrpc.UpdateQuoteResponse, error) {
	var fieldErrors []*quoteGrpc.ValidationError

	if req.QuoteId == "" {
		fieldErrors = append(fieldErrors, sqlutil.Required("quote_id"))
	}
	if req.UserId == "" {
		fieldErrors = append(fieldErrors, sqlutil.Required("user_id"))
	}
	if req.Name == "" {
		fieldErrors = append(fieldErrors, sqlutil.Required("name"))
	}

	if len(fieldErrors) > 0 {
		return &quoteGrpc.UpdateQuoteResponse{Success: false, Code: codes.InvalidInput, ValidationErrors: fieldErrors}, nil
	}

	if code, ok := EditableForUser(ctx, db, req.QuoteId, req.UserId); !ok {
		return &quoteGrpc.UpdateQuoteResponse{Success: false, Code: code}, nil
	}

	_, err := db.ExecContext(ctx,
		`UPDATE quotes SET
		    name            = $3,
		    client_id       = COALESCE(NULLIF($4, '')::TEXT, client_id),
		    address_id      = COALESCE(NULLIF($5, 0)::INT,  address_id),
		    user_address_id = COALESCE(NULLIF($6, 0)::INT,  user_address_id),
		    valid_until     = CASE WHEN $7 = '' THEN valid_until ELSE $7::DATE END,
		    payment_terms   = CASE WHEN $8 = '' THEN payment_terms ELSE $8 END,
		    state           = CASE WHEN state = 'negociation' THEN 'draft' ELSE state END,
		    updated_at      = NOW()
		 WHERE quote_id = $1 AND user_id = $2`,
		req.QuoteId, req.UserId, req.Name, req.ClientId, req.AddressId, req.UserAddressId,
		req.ValidUntil, req.PaymentTerms,
	)
	if err != nil {
		return &quoteGrpc.UpdateQuoteResponse{Success: false, Code: codes.InternalError}, err
	}

	return &quoteGrpc.UpdateQuoteResponse{Success: true, Code: codes.Success}, nil
}
