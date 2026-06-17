package fee

import (
	"context"
	"database/sql"

	"project-devis-quote/actions/codes"
	quoteGrpc "project-devis-quote/services/grpc"
)

func Get(ctx context.Context, db *sql.DB, req *quoteGrpc.GetFeeRequest) (*quoteGrpc.GetFeeResponse, error) {
	if req.FeeId == "" || req.UserId == "" {
		return &quoteGrpc.GetFeeResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	row := db.QueryRowContext(ctx,
		`SELECT `+selectColumns+` FROM fees WHERE fee_id=$1 AND user_id=$2`,
		req.FeeId, req.UserId,
	)
	f, err := scanFee(row)
	if err == sql.ErrNoRows {
		return &quoteGrpc.GetFeeResponse{Success: false, Code: codes.NotFound}, nil
	}
	if err != nil {
		return &quoteGrpc.GetFeeResponse{Success: false, Code: codes.InternalError}, err
	}

	return &quoteGrpc.GetFeeResponse{Success: true, Code: codes.Success, Fee: f}, nil
}
