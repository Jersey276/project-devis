package country

import (
	"context"
	"database/sql"

	usersGrpc "project-devis-users/services/grpc"
)

func List(ctx context.Context, db *sql.DB, _ *usersGrpc.ListCountriesRequest) (*usersGrpc.ListCountriesResponse, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, code, name FROM countries ORDER BY name")
	if err != nil {
		return &usersGrpc.ListCountriesResponse{Success: false, Code: codeInternalError}, err
	}
	defer rows.Close()

	var countries []*usersGrpc.Country
	for rows.Next() {
		var c usersGrpc.Country
		if err := rows.Scan(&c.Id, &c.Code, &c.Name); err != nil {
			return &usersGrpc.ListCountriesResponse{Success: false, Code: codeInternalError}, err
		}
		countries = append(countries, &c)
	}

	return &usersGrpc.ListCountriesResponse{Success: true, Code: codeSuccess, Countries: countries}, nil
}
