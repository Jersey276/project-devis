package country_group

import (
	"context"
	"database/sql"

	usersGrpc "project-devis-users/services/grpc"
)

func List(ctx context.Context, db *sql.DB, _ *usersGrpc.ListCountryGroupsRequest) (*usersGrpc.ListCountryGroupsResponse, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, name FROM country_groups ORDER BY name")
	if err != nil {
		return &usersGrpc.ListCountryGroupsResponse{Success: false, Code: codeInternalError}, err
	}
	defer rows.Close()

	var groups []*usersGrpc.CountryGroup
	for rows.Next() {
		var g usersGrpc.CountryGroup
		if err := rows.Scan(&g.Id, &g.Name); err != nil {
			return &usersGrpc.ListCountryGroupsResponse{Success: false, Code: codeInternalError}, err
		}
		countries, err := fetchCountries(ctx, db, g.Id)
		if err != nil {
			return &usersGrpc.ListCountryGroupsResponse{Success: false, Code: codeInternalError}, err
		}
		taxes, err := fetchTaxes(ctx, db, g.Id)
		if err != nil {
			return &usersGrpc.ListCountryGroupsResponse{Success: false, Code: codeInternalError}, err
		}
		g.Countries = countries
		g.Taxes = taxes
		groups = append(groups, &g)
	}

	return &usersGrpc.ListCountryGroupsResponse{Success: true, Code: codeSuccess, CountryGroups: groups}, nil
}
