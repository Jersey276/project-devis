package fee

import (
	"context"
	"database/sql"

	"project-devis-quote/actions/codes"
	quoteGrpc "project-devis-quote/services/grpc"
)

func List(ctx context.Context, db *sql.DB, req *quoteGrpc.ListFeesRequest) (*quoteGrpc.ListFeesResponse, error) {
	if req.UserId == "" {
		return &quoteGrpc.ListFeesResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	query := `SELECT ` + selectColumns + ` FROM fees WHERE user_id=$1`
	if !req.IncludeArchived {
		query += ` AND archived_at IS NULL`
	}
	query += ` ORDER BY created_at DESC`

	rows, err := db.QueryContext(ctx, query, req.UserId)
	if err != nil {
		return &quoteGrpc.ListFeesResponse{Success: false, Code: codes.InternalError}, err
	}
	defer rows.Close()

	var fees []*quoteGrpc.Fee
	for rows.Next() {
		f, err := scanFee(rows)
		if err != nil {
			return &quoteGrpc.ListFeesResponse{Success: false, Code: codes.InternalError}, err
		}
		fees = append(fees, f)
	}
	if err := rows.Err(); err != nil {
		return &quoteGrpc.ListFeesResponse{Success: false, Code: codes.InternalError}, err
	}

	return &quoteGrpc.ListFeesResponse{Success: true, Code: codes.Success, Fees: fees}, nil
}
