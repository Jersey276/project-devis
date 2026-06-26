package quote

import (
	"context"
	"database/sql"

	"project-devis-quote/actions/codes"
	quoteGrpc "project-devis-quote/services/grpc"
)

// Refuse moves a quote to 'refused'. Allowed from 'negociation' only.
// Called by the linked customer, not the provider.
func Refuse(ctx context.Context, db *sql.DB, req *quoteGrpc.RefuseQuoteRequest) (*quoteGrpc.GenericResponse, error) {
	if req.QuoteId == "" || req.UserId == "" {
		return &quoteGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}
	return setNegociationFinalState(ctx, db, req.QuoteId, req.UserId, StateRefused)
}
