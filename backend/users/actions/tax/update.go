package tax

import (
	"context"
	"database/sql"

	usersGrpc "project-devis-users/services/grpc"
)

func Update(ctx context.Context, db *sql.DB, req *usersGrpc.UpdateTaxRequest) (*usersGrpc.UpdateTaxResponse, error) {
	if req.TaxId == 0 {
		return &usersGrpc.UpdateTaxResponse{Success: false, Code: codeInvalidInput}, nil
	}
	if req.Rate != "" {
		if err := validateRate(req.Rate); err != nil {
			return &usersGrpc.UpdateTaxResponse{Success: false, Code: codeInvalidInput}, nil
		}
	}

	res, err := db.ExecContext(ctx,
		`UPDATE taxes SET name=COALESCE(NULLIF($1,''),name), rate=COALESCE(NULLIF($2,'')::DECIMAL,rate) WHERE id=$3`,
		req.Name, req.Rate, req.TaxId,
	)
	if err != nil {
		return &usersGrpc.UpdateTaxResponse{Success: false, Code: codeInternalError}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &usersGrpc.UpdateTaxResponse{Success: false, Code: codeNotFound}, nil
	}

	return &usersGrpc.UpdateTaxResponse{Success: true, Code: codeSuccess}, nil
}
