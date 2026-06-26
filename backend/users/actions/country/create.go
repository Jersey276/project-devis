package country

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
	"project-devis-users/actions/codes"
	"project-devis-users/actions/sqlutil"
	usersGrpc "project-devis-users/services/grpc"
)

func Create(ctx context.Context, db *sql.DB, req *usersGrpc.CreateCountryRequest) (*usersGrpc.CreateCountryResponse, error) {
	var fieldErrors []*usersGrpc.ValidationError

	if len(req.Code) != 2 {
		fieldErrors = append(fieldErrors, sqlutil.Invalid("code", "Doit être un code ISO à 2 caractères."))
	}
	if req.Name == "" {
		fieldErrors = append(fieldErrors, sqlutil.Required("name"))
	}

	if len(fieldErrors) > 0 {
		return &usersGrpc.CreateCountryResponse{Success: false, Code: codes.InvalidInput, ValidationErrors: fieldErrors}, nil
	}

	var countryID int32
	err := db.QueryRowContext(ctx,
		"INSERT INTO countries (code, name) VALUES ($1, $2) RETURNING id",
		req.Code, req.Name,
	).Scan(&countryID)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return &usersGrpc.CreateCountryResponse{Success: false, Code: codes.AlreadyExists}, nil
		}
		return &usersGrpc.CreateCountryResponse{Success: false, Code: codes.InternalError}, err
	}

	return &usersGrpc.CreateCountryResponse{Success: true, Code: codes.Success, CountryId: countryID}, nil
}
