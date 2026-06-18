package fee

import (
	"context"
	"database/sql"

	"project-devis-quote/actions/codes"
	quoteGrpc "project-devis-quote/services/grpc"
)

// Archive soft-deletes a fee. It deliberately does NOT touch existing quote
// lines: their snapshot survives so already-built quotes stay intact.
func Archive(ctx context.Context, db *sql.DB, req *quoteGrpc.ArchiveFeeRequest) (*quoteGrpc.GenericResponse, error) {
	if req.FeeId == "" || req.UserId == "" {
		return &quoteGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	res, err := db.ExecContext(ctx,
		`UPDATE fees SET archived_at=NOW(), updated_at=NOW()
		 WHERE fee_id=$1 AND user_id=$2 AND archived_at IS NULL`,
		req.FeeId, req.UserId,
	)
	if err != nil {
		return &quoteGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &quoteGrpc.GenericResponse{Success: false, Code: codes.NotFound}, nil
	}

	return &quoteGrpc.GenericResponse{Success: true, Code: codes.Success}, nil
}
