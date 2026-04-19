package country

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	usersGrpc "project-devis-users/services/grpc"
)

func Delete(ctx context.Context, db *sql.DB, req *usersGrpc.DeleteCountryRequest) (*usersGrpc.GenericResponse, error) {
	if req.CountryId == 0 {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	res, err := db.ExecContext(ctx, "DELETE FROM countries WHERE id=$1", req.CountryId)
	if err != nil {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.NotFound}, nil
	}

	return &usersGrpc.GenericResponse{Success: true, Code: codes.Success}, nil
}
