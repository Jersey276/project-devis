package address

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	"project-devis-users/actions/sqlutil"
	usersGrpc "project-devis-users/services/grpc"
)

func List(ctx context.Context, db *sql.DB, req *usersGrpc.ListAddressesRequest) (*usersGrpc.ListAddressesResponse, error) {
	ownerType, err := sqlutil.OwnerTypeToDBString(req.OwnerType)
	if err != nil || req.OwnerId == "" || req.AuthUserId == "" {
		return &usersGrpc.ListAddressesResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	rows, queryErr := db.QueryContext(ctx,
		`SELECT id, owner_type, owner_id, name, street, additional_street, city, zip_code, country_id, email, phone,
		        (archived_at IS NOT NULL)
		 FROM addresses
		 WHERE owner_type=$1 AND owner_id=$2 AND archived_at IS NULL
		   AND `+sqlutil.AddressAuthPredicate(3)+`
		 ORDER BY id`,
		ownerType, req.OwnerId, req.AuthUserId,
	)
	if queryErr != nil {
		return &usersGrpc.ListAddressesResponse{Success: false, Code: codes.InternalError}, queryErr
	}
	defer rows.Close()

	var addresses []*usersGrpc.Address
	for rows.Next() {
		var a usersGrpc.Address
		var ownerTypeDB string
		var addlStreet, email, phone sql.NullString
		if err := rows.Scan(&a.Id, &ownerTypeDB, &a.OwnerId, &a.Name, &a.Street, &addlStreet, &a.City, &a.ZipCode, &a.CountryId, &email, &phone, &a.Archived); err != nil {
			return &usersGrpc.ListAddressesResponse{Success: false, Code: codes.InternalError}, err
		}
		a.OwnerType, _ = sqlutil.OwnerTypeFromDBString(ownerTypeDB)
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
