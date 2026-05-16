package tax

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	usersGrpc "project-devis-users/services/grpc"
)

// Delete retires a tax: it sets superseded_at on the current row so it
// no longer appears in the available list. The row is preserved so that
// existing quote_lines referencing it keep their snapshot. The gRPC
// method name stays Delete to keep the contract stable; the gateway
// surfaces it as "Retirer".
func Delete(ctx context.Context, db *sql.DB, req *usersGrpc.DeleteTaxRequest) (*usersGrpc.GenericResponse, error) {
	if req.TaxId == 0 {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	res, err := db.ExecContext(ctx,
		"UPDATE taxes SET superseded_at=NOW(), is_default=FALSE WHERE id=$1 AND superseded_at IS NULL",
		req.TaxId,
	)
	if err != nil {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.NotFound}, nil
	}

	return &usersGrpc.GenericResponse{Success: true, Code: codes.Success}, nil
}
