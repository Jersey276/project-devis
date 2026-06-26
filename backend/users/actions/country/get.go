package country

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	usersGrpc "project-devis-users/services/grpc"
)

func Get(ctx context.Context, db *sql.DB, req *usersGrpc.GetCountryRequest) (*usersGrpc.GetCountryResponse, error) {
	if req.CountryId == 0 {
		return &usersGrpc.GetCountryResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	var c usersGrpc.Country
	err := db.QueryRowContext(ctx, "SELECT id, code, name, is_eu FROM countries WHERE id=$1", req.CountryId).
		Scan(&c.Id, &c.Code, &c.Name, &c.IsEu)
	if err == sql.ErrNoRows {
		return &usersGrpc.GetCountryResponse{Success: false, Code: codes.NotFound}, nil
	}
	if err != nil {
		return &usersGrpc.GetCountryResponse{Success: false, Code: codes.InternalError}, err
	}

	return &usersGrpc.GetCountryResponse{Success: true, Code: codes.Success, Country: &c}, nil
}
