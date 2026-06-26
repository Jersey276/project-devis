package quote

import (
	"context"
	"database/sql"

	"project-devis-quote/actions/codes"
	quoteGrpc "project-devis-quote/services/grpc"
)

// setNegociationFinalState moves a quote from 'negociation' to a terminal customer state.
func setNegociationFinalState(ctx context.Context, db *sql.DB, quoteID, userID, state string) (*quoteGrpc.GenericResponse, error) {
	res, err := db.ExecContext(ctx,
		`UPDATE quotes SET state=$3, updated_at=NOW()
		 WHERE quote_id=$1 AND user_id=$2 AND archived_at IS NULL
		   AND state='negociation'`,
		quoteID, userID, state,
	)
	if err != nil {
		return &quoteGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &quoteGrpc.GenericResponse{Success: false, Code: codes.NotFound}, nil
	}
	return &quoteGrpc.GenericResponse{Success: true, Code: codes.Success}, nil
}

// Accept moves a quote to 'accepted'. Allowed from 'negociation' only.
// Called by the linked customer, not the provider.
func Accept(ctx context.Context, db *sql.DB, req *quoteGrpc.AcceptQuoteRequest) (*quoteGrpc.GenericResponse, error) {
	if req.QuoteId == "" || req.UserId == "" {
		return &quoteGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}
	return setNegociationFinalState(ctx, db, req.QuoteId, req.UserId, StateAccepted)
}
