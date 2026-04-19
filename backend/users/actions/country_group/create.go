package country_group

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	usersGrpc "project-devis-users/services/grpc"
)

func Create(ctx context.Context, db *sql.DB, req *usersGrpc.CreateCountryGroupRequest) (*usersGrpc.CreateCountryGroupResponse, error) {
	if req.Name == "" {
		return &usersGrpc.CreateCountryGroupResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	var groupID int32
	err := db.QueryRowContext(ctx,
		"INSERT INTO country_groups (name) VALUES ($1) RETURNING id",
		req.Name,
	).Scan(&groupID)
	if err != nil {
		return &usersGrpc.CreateCountryGroupResponse{Success: false, Code: codes.InternalError}, err
	}

	return &usersGrpc.CreateCountryGroupResponse{Success: true, Code: codes.Success, CountryGroupId: groupID}, nil
}
