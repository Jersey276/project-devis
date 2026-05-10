package user

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	usersGrpc "project-devis-users/services/grpc"
)

// Delete removes the user and explicitly cleans up polymorphic addresses;
// the addresses table no longer carries a FK that can cascade.
//
// Order matters: client-owned addresses must be deleted before the user row,
// because clients.user_id has ON DELETE CASCADE — once the user is gone, the
// subquery that resolves owner_id from clients returns nothing and the
// addresses leak.
func Delete(ctx context.Context, db *sql.DB, req *usersGrpc.DeleteUserRequest) (*usersGrpc.GenericResponse, error) {
	if req.UserId == "" {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	defer tx.Rollback()

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM addresses WHERE owner_type='client' AND owner_id IN (SELECT client_id FROM clients WHERE user_id=$1)`,
		req.UserId,
	); err != nil {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}

	if _, err := tx.ExecContext(ctx,
		`DELETE FROM addresses WHERE owner_type='user' AND owner_id=$1`,
		req.UserId,
	); err != nil {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}

	res, err := tx.ExecContext(ctx, `DELETE FROM users WHERE user_id=$1`, req.UserId)
	if err != nil {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.NotFound}, nil
	}

	if err := tx.Commit(); err != nil {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}

	return &usersGrpc.GenericResponse{Success: true, Code: codes.Success}, nil
}
