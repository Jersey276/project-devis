package country_group

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	"project-devis-users/actions/tax"
	usersGrpc "project-devis-users/services/grpc"
)

func Get(ctx context.Context, db *sql.DB, req *usersGrpc.GetCountryGroupRequest) (*usersGrpc.GetCountryGroupResponse, error) {
	if req.CountryGroupId == 0 {
		return &usersGrpc.GetCountryGroupResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	var g usersGrpc.CountryGroup
	err := db.QueryRowContext(ctx, "SELECT id, name FROM country_groups WHERE id=$1", req.CountryGroupId).
		Scan(&g.Id, &g.Name)
	if err == sql.ErrNoRows {
		return &usersGrpc.GetCountryGroupResponse{Success: false, Code: codes.NotFound}, nil
	}
	if err != nil {
		return &usersGrpc.GetCountryGroupResponse{Success: false, Code: codes.InternalError}, err
	}

	countries, err := fetchCountries(ctx, db, g.Id)
	if err != nil {
		return &usersGrpc.GetCountryGroupResponse{Success: false, Code: codes.InternalError}, err
	}
	taxes, err := fetchTaxes(ctx, db, g.Id)
	if err != nil {
		return &usersGrpc.GetCountryGroupResponse{Success: false, Code: codes.InternalError}, err
	}
	g.Countries = countries
	g.Taxes = taxes

	return &usersGrpc.GetCountryGroupResponse{Success: true, Code: codes.Success, CountryGroup: &g}, nil
}

func fetchCountries(ctx context.Context, db *sql.DB, groupID int32) ([]*usersGrpc.Country, error) {
	rows, err := db.QueryContext(ctx,
		`SELECT c.id, c.code, c.name FROM countries c
		 JOIN country_group_countries cgc ON c.id = cgc.country_id
		 WHERE cgc.country_group_id = $1 ORDER BY c.name`,
		groupID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var countries []*usersGrpc.Country
	for rows.Next() {
		var c usersGrpc.Country
		if err := rows.Scan(&c.Id, &c.Code, &c.Name); err != nil {
			return nil, err
		}
		countries = append(countries, &c)
	}
	return countries, rows.Err()
}

func fetchTaxes(ctx context.Context, db *sql.DB, groupID int32) ([]*usersGrpc.Tax, error) {
	rows, err := db.QueryContext(ctx,
		"SELECT "+tax.Columns+" FROM taxes WHERE country_group_id=$1 AND superseded_at IS NULL ORDER BY name",
		groupID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return tax.ScanRows(rows)
}
