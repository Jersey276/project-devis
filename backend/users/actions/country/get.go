package country

import (
	"context"
	"database/sql"

	usersGrpc "project-devis-users/services/grpc"
)

func Get(ctx context.Context, db *sql.DB, req *usersGrpc.GetCountryRequest) (*usersGrpc.GetCountryResponse, error) {
	if req.CountryId == 0 {
		return &usersGrpc.GetCountryResponse{Success: false, Code: codeInvalidInput}, nil
	}

	var c usersGrpc.Country
	err := db.QueryRowContext(ctx, "SELECT id, code, name FROM countries WHERE id=$1", req.CountryId).
		Scan(&c.Id, &c.Code, &c.Name)
	if err == sql.ErrNoRows {
		return &usersGrpc.GetCountryResponse{Success: false, Code: codeNotFound}, nil
	}
	if err != nil {
		return &usersGrpc.GetCountryResponse{Success: false, Code: codeInternalError}, err
	}

	return &usersGrpc.GetCountryResponse{Success: true, Code: codeSuccess, Country: &c}, nil
}
