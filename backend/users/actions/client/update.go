package client

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	"project-devis-users/actions/sqlutil"
	usersGrpc "project-devis-users/services/grpc"
)

// Update is a FULL-REPLACE operation, not a partial update.
//
// Optional string fields (Email/Phone/Company/Siren/Vat) are written verbatim:
// an empty string clears the column to NULL via sqlutil.NullableStr. Callers
// MUST send the entire field set (typically by prefilling a form from a prior
// Get) — omitting a field will silently null it server-side.
//
// FirstName and LastName are always required.
func Update(ctx context.Context, db *sql.DB, req *usersGrpc.UpdateClientRequest) (*usersGrpc.UpdateClientResponse, error) {
	if req.ClientId == "" || req.UserId == "" || req.FirstName == "" || req.LastName == "" {
		return &usersGrpc.UpdateClientResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	res, err := db.ExecContext(ctx,
		`UPDATE clients SET first_name=$1, last_name=$2, email=$3, phone=$4,
		        company=$5, siren=$6, vat=$7, updated_at=NOW()
		 WHERE client_id=$8 AND user_id=$9 AND archived_at IS NULL`,
		req.FirstName, req.LastName,
		sqlutil.NullableStr(req.Email), sqlutil.NullableStr(req.Phone),
		sqlutil.NullableStr(req.Company), sqlutil.NullableStr(req.Siren), sqlutil.NullableStr(req.Vat),
		req.ClientId, req.UserId,
	)
	if err != nil {
		return &usersGrpc.UpdateClientResponse{Success: false, Code: codes.InternalError}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &usersGrpc.UpdateClientResponse{Success: false, Code: codes.NotFound}, nil
	}

	return &usersGrpc.UpdateClientResponse{Success: true, Code: codes.Success}, nil
}
