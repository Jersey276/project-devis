package tax

import (
	"context"
	"database/sql"

	usersGrpc "project-devis-users/services/grpc"
)

func Get(ctx context.Context, db *sql.DB, req *usersGrpc.GetTaxRequest) (*usersGrpc.GetTaxResponse, error) {
	if req.TaxId == 0 {
		return &usersGrpc.GetTaxResponse{Success: false, Code: codeInvalidInput}, nil
	}

	var t usersGrpc.Tax
	err := db.QueryRowContext(ctx,
		"SELECT id, name, rate::TEXT, country_group_id FROM taxes WHERE id=$1",
		req.TaxId,
	).Scan(&t.Id, &t.Name, &t.Rate, &t.CountryGroupId)
	if err == sql.ErrNoRows {
		return &usersGrpc.GetTaxResponse{Success: false, Code: codeNotFound}, nil
	}
	if err != nil {
		return &usersGrpc.GetTaxResponse{Success: false, Code: codeInternalError}, err
	}

	return &usersGrpc.GetTaxResponse{Success: true, Code: codeSuccess, Tax: &t}, nil
}
