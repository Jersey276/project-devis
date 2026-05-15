package tax

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
	"project-devis-users/actions/codes"
	"project-devis-users/actions/sqlutil"
	usersGrpc "project-devis-users/services/grpc"
)

// ListForUser returns the taxes available for the user's first registered
// address. The pattern (first address by id ascending) mirrors the export
// service's resolution of the prestataire country (see
// backend/export/actions/quote/export.go where addresses[0] is used).
//
// IncludeIds lets callers force specific tax rows into the response even
// if they are now superseded — used by the quote form to display the
// historical snapshot label of an old quote_line's tax_id.
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
		    AND taxes.superseded_at IS NULL
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

	if len(req.IncludeIds) > 0 {
		seen := make(map[int32]struct{}, len(taxes))
		for _, t := range taxes {
			seen[t.Id] = struct{}{}
		}
		var missing []int32
		for _, id := range req.IncludeIds {
			if _, ok := seen[id]; !ok {
				missing = append(missing, id)
			}
		}
		if len(missing) > 0 {
			extra, err := fetchByIDsForUser(ctx, db, req.UserId, missing)
			if err != nil {
				return &usersGrpc.ListTaxesResponse{Success: false, Code: codes.InternalError}, err
			}
			taxes = append(taxes, extra...)
		}
	}

	return &usersGrpc.ListTaxesResponse{Success: true, Code: codes.Success, Taxes: taxes}, nil
}

// fetchByIDsForUser loads taxes by id but restricts to country groups the
// user can legitimately reference (his first-address group). Without this
// restriction the include_ids parameter would let any caller enumerate the
// entire tax catalog across groups.
func fetchByIDsForUser(ctx context.Context, db *sql.DB, userID string, ids []int32) ([]*usersGrpc.Tax, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT `+Columns+`
		   FROM taxes
		  WHERE taxes.id = ANY($1)
		    AND taxes.country_group_id IN (
		        SELECT cgc.country_group_id
		          FROM country_group_countries cgc
		          JOIN addresses a ON a.country_id = cgc.country_id
		         WHERE a.owner_type=$2 AND a.owner_id=$3 AND a.archived_at IS NULL
		           AND a.id = (
		               SELECT MIN(id) FROM addresses
		                WHERE owner_type=$2 AND owner_id=$3 AND archived_at IS NULL
		           )
		    )
		  ORDER BY name, version DESC`,
		pq.Array(ids), sqlutil.OwnerTypeUser, userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return ScanRows(rows)
}
