package address

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	"project-devis-users/actions/sqlutil"
	usersGrpc "project-devis-users/services/grpc"
)

func Archive(ctx context.Context, db *sql.DB, req *usersGrpc.ArchiveAddressRequest) (*usersGrpc.GenericResponse, error) {
	ownerType, err := sqlutil.OwnerTypeToDBString(req.OwnerType)
	if err != nil || req.AddressId == 0 || req.OwnerId == "" || req.AuthUserId == "" {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	res, execErr := db.ExecContext(ctx,
		`UPDATE addresses SET archived_at=NOW()
		 WHERE id=$1 AND owner_type=$2 AND owner_id=$3 AND archived_at IS NULL
		   AND `+sqlutil.AddressAuthPredicate(4),
		req.AddressId, ownerType, req.OwnerId, req.AuthUserId,
	)
	if execErr != nil {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InternalError}, execErr
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.NotFound}, nil
	}

	return &usersGrpc.GenericResponse{Success: true, Code: codes.Success}, nil
}
