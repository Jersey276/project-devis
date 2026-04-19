package country_group

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	usersGrpc "project-devis-users/services/grpc"
)

func DetachCountry(ctx context.Context, db *sql.DB, req *usersGrpc.DetachCountryRequest) (*usersGrpc.GenericResponse, error) {
	if req.CountryGroupId == 0 || req.CountryId == 0 {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	res, err := db.ExecContext(ctx,
		"DELETE FROM country_group_countries WHERE country_group_id=$1 AND country_id=$2",
		req.CountryGroupId, req.CountryId,
	)
	if err != nil {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.NotFound}, nil
	}

	return &usersGrpc.GenericResponse{Success: true, Code: codes.Success}, nil
}
