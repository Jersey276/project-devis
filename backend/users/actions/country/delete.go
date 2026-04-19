package country

import (
	"context"
	"database/sql"

	usersGrpc "project-devis-users/services/grpc"
)

func Delete(ctx context.Context, db *sql.DB, req *usersGrpc.DeleteCountryRequest) (*usersGrpc.GenericResponse, error) {
	if req.CountryId == 0 {
		return &usersGrpc.GenericResponse{Success: false, Code: codeInvalidInput}, nil
	}

	res, err := db.ExecContext(ctx, "DELETE FROM countries WHERE id=$1", req.CountryId)
	if err != nil {
		return &usersGrpc.GenericResponse{Success: false, Code: codeInternalError}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &usersGrpc.GenericResponse{Success: false, Code: codeNotFound}, nil
	}

	return &usersGrpc.GenericResponse{Success: true, Code: codeSuccess}, nil
}
