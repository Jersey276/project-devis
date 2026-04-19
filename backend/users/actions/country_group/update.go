package country_group

import (
	"context"
	"database/sql"

	usersGrpc "project-devis-users/services/grpc"
)

func Update(ctx context.Context, db *sql.DB, req *usersGrpc.UpdateCountryGroupRequest) (*usersGrpc.UpdateCountryGroupResponse, error) {
	if req.CountryGroupId == 0 || req.Name == "" {
		return &usersGrpc.UpdateCountryGroupResponse{Success: false, Code: codeInvalidInput}, nil
	}

	res, err := db.ExecContext(ctx,
		"UPDATE country_groups SET name=$1 WHERE id=$2",
		req.Name, req.CountryGroupId,
	)
	if err != nil {
		return &usersGrpc.UpdateCountryGroupResponse{Success: false, Code: codeInternalError}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &usersGrpc.UpdateCountryGroupResponse{Success: false, Code: codeNotFound}, nil
	}

	return &usersGrpc.UpdateCountryGroupResponse{Success: true, Code: codeSuccess}, nil
}
