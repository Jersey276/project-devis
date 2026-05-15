package tax

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	"project-devis-users/actions/sqlutil"
	usersGrpc "project-devis-users/services/grpc"
)

// ListForUser returns the taxes available for the user's first registered
// address. The pattern (first address by id ascending) mirrors the export
// service's resolution of the prestataire country (see
// backend/export/actions/quote/export.go where addresses[0] is used).
func ListForUser(ctx context.Context, db *sql.DB, req *usersGrpc.ListTaxesForUserRequest) (*usersGrpc.ListTaxesResponse, error) {
	if req.UserId == "" {
		return &usersGrpc.ListTaxesResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	rows, err := db.QueryContext(ctx,
		`SELECT `+Columns+`
		   FROM taxes
		   JOIN country_group_countries cgc ON cgc.country_group_id = taxes.country_group_id
		   JOIN addresses a ON a.country_id = cgc.country_id
		  WHERE a.owner_type=$1 AND a.owner_id=$2 AND a.archived_at IS NULL
		    AND a.id = (
		        SELECT MIN(id) FROM addresses
		         WHERE owner_type=$1 AND owner_id=$2 AND archived_at IS NULL
		    )
		  ORDER BY name`,
		sqlutil.OwnerTypeUser, req.UserId,
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
