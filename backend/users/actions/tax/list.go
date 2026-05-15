package tax

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	usersGrpc "project-devis-users/services/grpc"
)

func List(ctx context.Context, db *sql.DB, req *usersGrpc.ListTaxesRequest) (*usersGrpc.ListTaxesResponse, error) {
	var rows *sql.Rows
	var err error

	if req.CountryGroupId != 0 {
		rows, err = db.QueryContext(ctx,
			"SELECT "+Columns+" FROM taxes WHERE country_group_id=$1 ORDER BY name",
			req.CountryGroupId,
		)
	} else {
		rows, err = db.QueryContext(ctx,
			"SELECT "+Columns+" FROM taxes ORDER BY name",
		)
	}
	if err != nil {
		return &usersGrpc.ListTaxesResponse{Success: false, Code: codes.InternalError}, err
	}
	defer rows.Close()

	taxes, err := ScanRows(rows)
	if err != nil {
		return &usersGrpc.ListTaxesResponse{Success: false, Code: codes.InternalError}, err
	}

	return &usersGrpc.ListTaxesResponse{Success: true, Code: codes.Success, Taxes: taxes}, nil
}
