package country_group

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	usersGrpc "project-devis-users/services/grpc"
)

func Update(ctx context.Context, db *sql.DB, req *usersGrpc.UpdateCountryGroupRequest) (*usersGrpc.UpdateCountryGroupResponse, error) {
	var fieldErrors []*usersGrpc.ValidationError

	if req.CountryGroupId == 0 {
		fieldErrors = append(fieldErrors, &usersGrpc.ValidationError{Field: "country_group_id", Message: "Champ requis."})
	}
	if req.Name == "" {
		fieldErrors = append(fieldErrors, &usersGrpc.ValidationError{Field: "name", Message: "Champ requis."})
	}

	if len(fieldErrors) > 0 {
		return &usersGrpc.UpdateCountryGroupResponse{Success: false, Code: codes.InvalidInput, ValidationErrors: fieldErrors}, nil
	}

	res, err := db.ExecContext(ctx,
		"UPDATE country_groups SET name=$1 WHERE id=$2",
		req.Name, req.CountryGroupId,
	)
	if err != nil {
		return &usersGrpc.UpdateCountryGroupResponse{Success: false, Code: codes.InternalError}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &usersGrpc.UpdateCountryGroupResponse{Success: false, Code: codes.NotFound}, nil
	}

	return &usersGrpc.UpdateCountryGroupResponse{Success: true, Code: codes.Success}, nil
}
