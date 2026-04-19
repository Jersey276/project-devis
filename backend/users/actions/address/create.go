package address

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	"project-devis-users/actions/sqlutil"
	usersGrpc "project-devis-users/services/grpc"
)

func Create(ctx context.Context, db *sql.DB, req *usersGrpc.CreateAddressRequest) (*usersGrpc.CreateAddressResponse, error) {
	if req.UserId == "" || req.Name == "" || req.Street == "" || req.City == "" || req.ZipCode == "" || req.CountryId == 0 {
		return &usersGrpc.CreateAddressResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	var addressID int32
	err := db.QueryRowContext(ctx,
		`INSERT INTO addresses (user_id, name, street, additional_street, city, zip_code, country_id, email, phone)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9) RETURNING id`,
		req.UserId, req.Name, req.Street, sqlutil.NullableStr(req.AdditionalStreet),
		req.City, req.ZipCode, req.CountryId, sqlutil.NullableStr(req.Email), sqlutil.NullableStr(req.Phone),
	).Scan(&addressID)
	if err != nil {
		return &usersGrpc.CreateAddressResponse{Success: false, Code: codes.InternalError}, err
	}

	return &usersGrpc.CreateAddressResponse{Success: true, Code: codes.Success, AddressId: addressID}, nil
}
