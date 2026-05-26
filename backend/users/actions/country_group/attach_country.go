package country_group

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	usersGrpc "project-devis-users/services/grpc"
)

func AttachCountry(ctx context.Context, db *sql.DB, req *usersGrpc.AttachCountryRequest) (*usersGrpc.GenericResponse, error) {
	if req.CountryGroupId == 0 || req.CountryId == 0 {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	_, err := db.ExecContext(ctx,
		"INSERT INTO country_group_countries (country_group_id, country_id) VALUES ($1,$2) ON CONFLICT DO NOTHING",
		req.CountryGroupId, req.CountryId,
	)
	if err != nil {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}

	return &usersGrpc.GenericResponse{Success: true, Code: codes.Success}, nil
}
