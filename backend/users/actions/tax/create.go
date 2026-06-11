package tax

import (
	"context"
	"database/sql"
	"fmt"
	"strconv"

	"project-devis-users/actions/codes"
	usersGrpc "project-devis-users/services/grpc"
)

func Create(ctx context.Context, db *sql.DB, req *usersGrpc.CreateTaxRequest) (*usersGrpc.CreateTaxResponse, error) {
	var fieldErrors []*usersGrpc.ValidationError

	if req.Name == "" {
		fieldErrors = append(fieldErrors, &usersGrpc.ValidationError{Field: "name", Message: "Champ requis."})
	}
	if req.Rate == "" {
		fieldErrors = append(fieldErrors, &usersGrpc.ValidationError{Field: "rate", Message: "Champ requis."})
	} else if err := validateRate(req.Rate); err != nil {
		fieldErrors = append(fieldErrors, &usersGrpc.ValidationError{Field: "rate", Message: "Taux invalide (0–999.99)."})
	}
	if req.CountryGroupId == 0 {
		fieldErrors = append(fieldErrors, &usersGrpc.ValidationError{Field: "country_group_id", Message: "Champ requis."})
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

func validateRate(rate string) error {
	v, err := strconv.ParseFloat(rate, 64)
	if err != nil {
		return fmt.Errorf("invalid rate format")
	}
	if v < 0 || v > 999.99 {
		return fmt.Errorf("rate out of range")
	}
	return nil
}
