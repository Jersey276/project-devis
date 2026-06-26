package tax

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	usersGrpc "project-devis-users/services/grpc"
)

func ListForCountry(ctx context.Context, db *sql.DB, req *usersGrpc.ListTaxesForCountryRequest) (*usersGrpc.ListTaxesResponse, error) {
	if req.CountryId == 0 {
		return &usersGrpc.ListTaxesResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	rows, err := db.QueryContext(ctx,
		`SELECT `+Columns+`
		   FROM taxes
		  WHERE taxes.country_group_id IN (
		      SELECT cgc.country_group_id
		        FROM country_group_countries cgc
		       WHERE cgc.country_id = $1
		  )
		    AND taxes.superseded_at IS NULL
		  ORDER BY name, version DESC`,
		req.CountryId,
	)
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
