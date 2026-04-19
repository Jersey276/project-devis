package address

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	usersGrpc "project-devis-users/services/grpc"
)

func Get(ctx context.Context, db *sql.DB, req *usersGrpc.GetAddressRequest) (*usersGrpc.GetAddressResponse, error) {
	if req.AddressId == 0 || req.UserId == "" {
		return &usersGrpc.GetAddressResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	var a usersGrpc.Address
	var addlStreet, email, phone sql.NullString
	err := db.QueryRowContext(ctx,
		`SELECT id, user_id, name, street, additional_street, city, zip_code, country_id, email, phone,
		        (archived_at IS NOT NULL)
		 FROM addresses WHERE id=$1 AND user_id=$2`,
		req.AddressId, req.UserId,
	).Scan(&a.Id, &a.UserId, &a.Name, &a.Street, &addlStreet, &a.City, &a.ZipCode, &a.CountryId, &email, &phone, &a.Archived)
	if err == sql.ErrNoRows {
		return &usersGrpc.GetAddressResponse{Success: false, Code: codes.NotFound}, nil
	}
	if err != nil {
		return &usersGrpc.GetAddressResponse{Success: false, Code: codes.InternalError}, err
	}

	a.AdditionalStreet = addlStreet.String
	a.Email = email.String
	a.Phone = phone.String

	return &usersGrpc.GetAddressResponse{Success: true, Code: codes.Success, Address: &a}, nil
}
