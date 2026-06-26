package tax

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	"project-devis-users/actions/sqlutil"
	usersGrpc "project-devis-users/services/grpc"
)

func Create(ctx context.Context, db *sql.DB, req *usersGrpc.CreateTaxRequest) (*usersGrpc.CreateTaxResponse, error) {
	var fieldErrors []*usersGrpc.ValidationError

	if req.Name == "" {
		fieldErrors = append(fieldErrors, sqlutil.Required("name"))
	}
	if req.Rate == "" {
		fieldErrors = append(fieldErrors, sqlutil.Required("rate"))
	} else if err := sqlutil.ValidateRate(req.Rate); err != nil {
		fieldErrors = append(fieldErrors, sqlutil.Invalid("rate", "Taux invalide (0–999.99)."))
	}
	if req.CountryGroupId == 0 {
		fieldErrors = append(fieldErrors, sqlutil.Required("country_group_id"))
	}

	if len(fieldErrors) > 0 {
		return &usersGrpc.CreateTaxResponse{Success: false, Code: codes.InvalidInput, ValidationErrors: fieldErrors}, nil
	}

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return &usersGrpc.CreateTaxResponse{Success: false, Code: codes.InternalError}, err
	}
	defer tx.Rollback()

	if req.IsDefault {
		if err := clearDefaultInGroup(ctx, tx, req.CountryGroupId, 0); err != nil {
			return &usersGrpc.CreateTaxResponse{Success: false, Code: codes.InternalError}, err
		}
	}

	var taxID int32
	if err := tx.QueryRowContext(ctx,
		"INSERT INTO taxes (name, rate, country_group_id, is_default) VALUES ($1, $2, $3, $4) RETURNING id",
		req.Name, req.Rate, req.CountryGroupId, req.IsDefault,
	).Scan(&taxID); err != nil {
		return &usersGrpc.CreateTaxResponse{Success: false, Code: codes.InternalError}, err
	}

	if err := tx.Commit(); err != nil {
		return &usersGrpc.CreateTaxResponse{Success: false, Code: codes.InternalError}, err
	}

	return &usersGrpc.CreateTaxResponse{Success: true, Code: codes.Success, TaxId: taxID}, nil
}
