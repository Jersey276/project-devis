package tax

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	usersGrpc "project-devis-users/services/grpc"
)

func Update(ctx context.Context, db *sql.DB, req *usersGrpc.UpdateTaxRequest) (*usersGrpc.UpdateTaxResponse, error) {
	if req.TaxId == 0 {
		return &usersGrpc.UpdateTaxResponse{Success: false, Code: codes.InvalidInput}, nil
	}
	if req.Rate != "" {
		if err := validateRate(req.Rate); err != nil {
			return &usersGrpc.UpdateTaxResponse{Success: false, Code: codes.InvalidInput}, nil
		}
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return &usersGrpc.UpdateTaxResponse{Success: false, Code: codes.InternalError}, err
	}
	defer tx.Rollback()

	var groupID int32
	if err := tx.QueryRowContext(ctx,
		"SELECT country_group_id FROM taxes WHERE id=$1",
		req.TaxId,
	).Scan(&groupID); err != nil {
		if err == sql.ErrNoRows {
			return &usersGrpc.UpdateTaxResponse{Success: false, Code: codes.NotFound}, nil
		}
		return &usersGrpc.UpdateTaxResponse{Success: false, Code: codes.InternalError}, err
	}

	if req.IsDefault {
		if _, err := tx.ExecContext(ctx,
			"UPDATE taxes SET is_default=FALSE WHERE country_group_id=$1 AND id<>$2 AND is_default=TRUE",
			groupID, req.TaxId,
		); err != nil {
			return &usersGrpc.UpdateTaxResponse{Success: false, Code: codes.InternalError}, err
		}
	}

	res, err := tx.ExecContext(ctx,
		`UPDATE taxes
		    SET name=COALESCE(NULLIF($1,''),name),
		        rate=COALESCE(NULLIF($2,'')::DECIMAL,rate),
		        is_default=$3
		  WHERE id=$4`,
		req.Name, req.Rate, req.IsDefault, req.TaxId,
	)
	if err != nil {
		return &usersGrpc.UpdateTaxResponse{Success: false, Code: codes.InternalError}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &usersGrpc.UpdateTaxResponse{Success: false, Code: codes.NotFound}, nil
	}

	if err := tx.Commit(); err != nil {
		return &usersGrpc.UpdateTaxResponse{Success: false, Code: codes.InternalError}, err
	}

	return &usersGrpc.UpdateTaxResponse{Success: true, Code: codes.Success}, nil
}
