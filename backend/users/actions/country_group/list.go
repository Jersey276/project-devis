package country_group

import (
	"context"
	"database/sql"

	"github.com/lib/pq"
	"project-devis-users/actions/codes"
	"project-devis-users/actions/tax"
	usersGrpc "project-devis-users/services/grpc"
)

func List(ctx context.Context, db *sql.DB, _ *usersGrpc.ListCountryGroupsRequest) (*usersGrpc.ListCountryGroupsResponse, error) {
	rows, err := db.QueryContext(ctx, "SELECT id, name FROM country_groups ORDER BY name")
	if err != nil {
		return &usersGrpc.ListCountryGroupsResponse{Success: false, Code: codes.InternalError}, err
	}
	defer rows.Close()

	var groups []*usersGrpc.CountryGroup
	for rows.Next() {
		var g usersGrpc.CountryGroup
		if err := rows.Scan(&g.Id, &g.Name); err != nil {
			return &usersGrpc.ListCountryGroupsResponse{Success: false, Code: codes.InternalError}, err
		}
		groups = append(groups, &g)
	}
	if err := rows.Err(); err != nil {
		return &usersGrpc.ListCountryGroupsResponse{Success: false, Code: codes.InternalError}, err
	}

	if len(groups) == 0 {
		return &usersGrpc.ListCountryGroupsResponse{Success: true, Code: codes.Success, CountryGroups: groups}, nil
	}

	groupIDs := make([]int32, len(groups))
	for i, g := range groups {
		groupIDs[i] = g.Id
	}

	countriesByGroup, err := fetchAllCountries(ctx, db, groupIDs)
	if err != nil {
		return &usersGrpc.ListCountryGroupsResponse{Success: false, Code: codes.InternalError}, err
	}
	taxesByGroup, err := fetchAllTaxes(ctx, db, groupIDs)
	if err != nil {
		return &usersGrpc.ListCountryGroupsResponse{Success: false, Code: codes.InternalError}, err
	}

	for _, g := range groups {
		g.Countries = countriesByGroup[g.Id]
		g.Taxes = taxesByGroup[g.Id]
	}

	return &usersGrpc.ListCountryGroupsResponse{Success: true, Code: codes.Success, CountryGroups: groups}, nil
}

func fetchAllCountries(ctx context.Context, db *sql.DB, groupIDs []int32) (map[int32][]*usersGrpc.Country, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT c.id, c.code, c.name, cgc.country_group_id
		 FROM countries c
		 JOIN country_group_countries cgc ON c.id = cgc.country_id
		 WHERE cgc.country_group_id = ANY($1) ORDER BY c.name`,
		pq.Array(groupIDs),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[int32][]*usersGrpc.Country)
	for rows.Next() {
		var c usersGrpc.Country
		var groupID int32
		if err := rows.Scan(&c.Id, &c.Code, &c.Name, &groupID); err != nil {
			return nil, err
		}
		result[groupID] = append(result[groupID], &c)
	}
	return result, rows.Err()
}

func fetchAllTaxes(ctx context.Context, db *sql.DB, groupIDs []int32) (map[int32][]*usersGrpc.Tax, error) {
	rows, err := db.QueryContext(ctx,
		"SELECT "+tax.Columns+" FROM taxes WHERE country_group_id = ANY($1) ORDER BY name",
		pq.Array(groupIDs),
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	taxes, err := tax.ScanRows(rows)
	if err != nil {
		return nil, err
	}
	result := make(map[int32][]*usersGrpc.Tax)
	for _, t := range taxes {
		result[t.CountryGroupId] = append(result[t.CountryGroupId], t)
	}
	return result, nil
}
