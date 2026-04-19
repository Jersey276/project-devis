package country_group

import (
	"context"
	"database/sql"

	usersGrpc "project-devis-users/services/grpc"
)

func Delete(ctx context.Context, db *sql.DB, req *usersGrpc.DeleteCountryGroupRequest) (*usersGrpc.GenericResponse, error) {
	if req.CountryGroupId == 0 {
		return &usersGrpc.GenericResponse{Success: false, Code: codeInvalidInput}, nil
	}

	res, err := db.ExecContext(ctx, "DELETE FROM country_groups WHERE id=$1", req.CountryGroupId)
	if err != nil {
		return &usersGrpc.GenericResponse{Success: false, Code: codeInternalError}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &usersGrpc.GenericResponse{Success: false, Code: codeNotFound}, nil
	}

	return &usersGrpc.GenericResponse{Success: true, Code: codeSuccess}, nil
}
