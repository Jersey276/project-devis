package country

import (
	"context"
	"database/sql"

	usersGrpc "project-devis-users/services/grpc"
)

func Create(ctx context.Context, db *sql.DB, req *usersGrpc.CreateCountryRequest) (*usersGrpc.CreateCountryResponse, error) {
	if len(req.Code) != 2 || req.Name == "" {
		return &usersGrpc.CreateCountryResponse{Success: false, Code: codeInvalidInput}, nil
	}

	var existing int
	err := db.QueryRowContext(ctx, "SELECT id FROM countries WHERE code=$1", req.Code).Scan(&existing)
	if err == nil {
		return &usersGrpc.CreateCountryResponse{Success: false, Code: codeAlreadyExists}, nil
	}
	if err != sql.ErrNoRows {
		return &usersGrpc.CreateCountryResponse{Success: false, Code: codeInternalError}, err
	}

	var countryID int32
	err = db.QueryRowContext(ctx,
		"INSERT INTO countries (code, name) VALUES ($1, $2) RETURNING id",
		req.Code, req.Name,
	).Scan(&countryID)
	if err != nil {
		return &usersGrpc.CreateCountryResponse{Success: false, Code: codeInternalError}, err
	}

	return &usersGrpc.CreateCountryResponse{Success: true, Code: codeSuccess, CountryId: countryID}, nil
}
