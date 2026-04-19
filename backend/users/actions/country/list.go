package country

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	usersGrpc "project-devis-users/services/grpc"
)

func List(ctx context.Context, db *sql.DB, _ *usersGrpc.ListCountriesRequest) (*usersGrpc.ListCountriesResponse, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, code, name FROM countries ORDER BY name")
	if err != nil {
		return &usersGrpc.ListCountriesResponse{Success: false, Code: codes.InternalError}, err
	}
	defer rows.Close()

	var countries []*usersGrpc.Country
	for rows.Next() {
		var c usersGrpc.Country
		if err := rows.Scan(&c.Id, &c.Code, &c.Name); err != nil {
			return &usersGrpc.ListCountriesResponse{Success: false, Code: codes.InternalError}, err
		}
		countries = append(countries, &c)
	}
	if err := rows.Err(); err != nil {
		return &usersGrpc.ListCountriesResponse{Success: false, Code: codes.InternalError}, err
	}

	return &usersGrpc.ListCountriesResponse{Success: true, Code: codes.Success, Countries: countries}, nil
}
