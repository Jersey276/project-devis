package quote

import (
	"context"
	"database/sql"

	"project-devis-quote/actions/codes"
	quoteGrpc "project-devis-quote/services/grpc"
)

// Trash hard-deletes every archived quote owned by the user (cascades to lines).
func Trash(ctx context.Context, db *sql.DB, req *quoteGrpc.TrashQuotesRequest) (*quoteGrpc.GenericResponse, error) {
	if req.UserId == "" {
		return &quoteGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	if _, err := db.ExecContext(ctx,
		`DELETE FROM quotes WHERE user_id=$1 AND archived_at IS NOT NULL`,
		req.UserId,
	); err != nil {
		return &quoteGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}

	return &quoteGrpc.GenericResponse{Success: true, Code: codes.Success}, nil
}
