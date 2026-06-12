package country

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	usersGrpc "project-devis-users/services/grpc"
)

func Update(ctx context.Context, db *sql.DB, req *usersGrpc.UpdateCountryRequest) (*usersGrpc.UpdateCountryResponse, error) {
	var fieldErrors []*usersGrpc.ValidationError

	if req.CountryId == 0 {
		fieldErrors = append(fieldErrors, &usersGrpc.ValidationError{Field: "country_id", Message: "Champ requis."})
	}
	if len(req.Code) > 0 && len(req.Code) != 2 {
		fieldErrors = append(fieldErrors, &usersGrpc.ValidationError{Field: "code", Message: "Doit être un code ISO à 2 caractères."})
	}

	if len(fieldErrors) > 0 {
		return &usersGrpc.UpdateCountryResponse{Success: false, Code: codes.InvalidInput, ValidationErrors: fieldErrors}, nil
	}

	res, err := db.ExecContext(ctx,
		"UPDATE countries SET code=COALESCE(NULLIF($1,''),code), name=COALESCE(NULLIF($2,''),name) WHERE id=$3",
		req.Code, req.Name, req.CountryId,
	)
	if err != nil {
		return &usersGrpc.UpdateCountryResponse{Success: false, Code: codes.InternalError}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &usersGrpc.UpdateCountryResponse{Success: false, Code: codes.NotFound}, nil
	}

	return &usersGrpc.UpdateCountryResponse{Success: true, Code: codes.Success}, nil
}
