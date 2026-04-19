package tax

import (
	"context"
	"database/sql"

	usersGrpc "project-devis-users/services/grpc"
)

func List(ctx context.Context, db *sql.DB, req *usersGrpc.ListTaxesRequest) (*usersGrpc.ListTaxesResponse, error) {
	var rows *sql.Rows
	var err error

	if req.CountryGroupId != 0 {
		rows, err = db.QueryContext(ctx,
			"SELECT id, name, rate::TEXT, country_group_id FROM taxes WHERE country_group_id=$1 ORDER BY name",
			req.CountryGroupId,
		)
	} else {
		rows, err = db.QueryContext(ctx,
			"SELECT id, name, rate::TEXT, country_group_id FROM taxes ORDER BY name",
		)
	}
	if err != nil {
		return &usersGrpc.ListTaxesResponse{Success: false, Code: codeInternalError}, err
	}
	defer rows.Close()

	var taxes []*usersGrpc.Tax
	for rows.Next() {
		var t usersGrpc.Tax
		if err := rows.Scan(&t.Id, &t.Name, &t.Rate, &t.CountryGroupId); err != nil {
			return &usersGrpc.ListTaxesResponse{Success: false, Code: codeInternalError}, err
		}
		taxes = append(taxes, &t)
	}

	return &usersGrpc.ListTaxesResponse{Success: true, Code: codeSuccess, Taxes: taxes}, nil
}
