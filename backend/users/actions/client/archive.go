package client

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	usersGrpc "project-devis-users/services/grpc"
)

func Archive(ctx context.Context, db *sql.DB, req *usersGrpc.ArchiveClientRequest) (*usersGrpc.GenericResponse, error) {
	if req.ClientId == "" || req.UserId == "" {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	res, err := db.ExecContext(ctx,
		"UPDATE clients SET archived_at=NOW() WHERE client_id=$1 AND user_id=$2 AND archived_at IS NULL",
		req.ClientId, req.UserId,
	)
	if err != nil {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.NotFound}, nil
	}

	return &usersGrpc.GenericResponse{Success: true, Code: codes.Success}, nil
}
