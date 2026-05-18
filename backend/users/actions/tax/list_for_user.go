package tax

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	"project-devis-users/actions/sqlutil"
	usersGrpc "project-devis-users/services/grpc"

	"github.com/lib/pq"
)

func ListForUser(ctx context.Context, db *sql.DB, req *usersGrpc.ListTaxesForUserRequest) (*usersGrpc.ListTaxesResponse, error) {
	if req.UserId == "" {
		return &usersGrpc.ListTaxesResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	rows, err := db.QueryContext(ctx,
		`SELECT `+Columns+`
		   FROM taxes
		  WHERE taxes.country_group_id IN (
		      SELECT cgc.country_group_id
		        FROM country_group_countries cgc
		        JOIN addresses a ON a.country_id = cgc.country_id
		       WHERE a.owner_type=$1 AND a.owner_id=$2 AND a.archived_at IS NULL
		         AND a.id = CASE
		             WHEN $4::INT > 0 THEN $4::INT
		             ELSE (
		                 SELECT MIN(id) FROM addresses
		                  WHERE owner_type=$1 AND owner_id=$2 AND archived_at IS NULL
		             )
		         END
		  )
		    AND (taxes.superseded_at IS NULL OR taxes.id = ANY($3))
		  ORDER BY name, version DESC`,
		sqlutil.OwnerTypeUser, req.UserId, pq.Array(req.IncludeIds), req.AddressId,
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
