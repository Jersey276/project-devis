package tax

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"project-devis-users/actions/codes"
	usersGrpc "project-devis-users/services/grpc"
)

func Create(ctx context.Context, db *sql.DB, req *usersGrpc.CreateTaxRequest) (*usersGrpc.CreateTaxResponse, error) {
	if req.Name == "" || req.Rate == "" || req.CountryGroupId == 0 {
		return &usersGrpc.CreateTaxResponse{Success: false, Code: codes.InvalidInput}, nil
	}
	if err := validateRate(req.Rate); err != nil {
		return &usersGrpc.CreateTaxResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	var taxID int32
	err := db.QueryRowContext(ctx,
		"INSERT INTO taxes (name, rate, country_group_id) VALUES ($1, $2, $3) RETURNING id",
		req.Name, req.Rate, req.CountryGroupId,
	).Scan(&taxID)
	if err != nil {
		return &usersGrpc.CreateTaxResponse{Success: false, Code: codes.InternalError}, err
	}

	return &usersGrpc.CreateTaxResponse{Success: true, Code: codes.Success, TaxId: taxID}, nil
}

func validateRate(rate string) error {
	v, err := strconv.ParseFloat(rate, 64)
	if err != nil {
		return fmt.Errorf("invalid rate format")
	}
	if v < 0 || v > 999.99 {
		return fmt.Errorf("rate out of range")
	}
	return nil
}
