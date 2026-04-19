package address

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	usersGrpc "project-devis-users/services/grpc"
)

func List(ctx context.Context, db *sql.DB, req *usersGrpc.ListAddressesRequest) (*usersGrpc.ListAddressesResponse, error) {
	if req.UserId == "" {
		return &usersGrpc.ListAddressesResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	rows, err := db.QueryContext(ctx,
		`SELECT id, user_id, name, street, additional_street, city, zip_code, country_id, email, phone,
		        (archived_at IS NOT NULL)
		 FROM addresses WHERE user_id=$1 AND archived_at IS NULL ORDER BY id`,
		req.UserId,
	)
	if err != nil {
		return &usersGrpc.ListAddressesResponse{Success: false, Code: codes.InternalError}, err
	}
	defer rows.Close()

	var addresses []*usersGrpc.Address
	for rows.Next() {
		var a usersGrpc.Address
		var addlStreet, email, phone sql.NullString
		if err := rows.Scan(&a.Id, &a.UserId, &a.Name, &a.Street, &addlStreet, &a.City, &a.ZipCode, &a.CountryId, &email, &phone, &a.Archived); err != nil {
			return &usersGrpc.ListAddressesResponse{Success: false, Code: codes.InternalError}, err
		}
		a.AdditionalStreet = addlStreet.String
		a.Email = email.String
		a.Phone = phone.String
		addresses = append(addresses, &a)
	}
	if err := rows.Err(); err != nil {
		return &usersGrpc.ListAddressesResponse{Success: false, Code: codes.InternalError}, err
	}

	return &usersGrpc.ListAddressesResponse{Success: true, Code: codes.Success, Addresses: addresses}, nil
}
