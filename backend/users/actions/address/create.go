package address

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	"project-devis-users/actions/sqlutil"
	usersGrpc "project-devis-users/services/grpc"
)

func Create(ctx context.Context, db *sql.DB, req *usersGrpc.CreateAddressRequest) (*usersGrpc.CreateAddressResponse, error) {
	var fieldErrors []*usersGrpc.ValidationError

	ownerType, err := sqlutil.OwnerTypeToDBString(req.OwnerType)
	if err != nil {
		fieldErrors = append(fieldErrors, sqlutil.Invalid("owner_type", "Type de propriétaire invalide."))
	}
	if req.OwnerId == "" {
		fieldErrors = append(fieldErrors, sqlutil.Required("owner_id"))
	}
	if req.AuthUserId == "" {
		fieldErrors = append(fieldErrors, sqlutil.Required("auth_user_id"))
	}
	if req.Name == "" {
		fieldErrors = append(fieldErrors, sqlutil.Required("name"))
	}
	if req.Street == "" {
		fieldErrors = append(fieldErrors, sqlutil.Required("street"))
	}
	if req.City == "" {
		fieldErrors = append(fieldErrors, sqlutil.Required("city"))
	}
	if req.ZipCode == "" {
		fieldErrors = append(fieldErrors, sqlutil.Required("zip_code"))
	}
	if req.CountryId == 0 {
		fieldErrors = append(fieldErrors, sqlutil.Required("country_id"))
	}

	if len(fieldErrors) > 0 {
		return &usersGrpc.CreateAddressResponse{Success: false, Code: codes.InvalidInput, ValidationErrors: fieldErrors}, nil
	}

	// INSERT-SELECT with the auth predicate baked in: zero rows inserted means
	// the (owner_type, owner_id) the caller supplied is not reachable from the
	// authenticated user — surface as NotFound rather than swallowing as success.
	var addressID int32
	queryErr := db.QueryRowContext(ctx,
		`INSERT INTO addresses (owner_type, owner_id, name, street, additional_street, city, zip_code, country_id, email, phone)
		 SELECT $1, $2, $3, $4, $5, $6, $7, $8, $9, $10
		 WHERE
		   ($1 = 'user'   AND $2 = $11)
		   OR
		   ($1 = 'client' AND $2 IN (
		     SELECT client_id FROM clients
		     WHERE user_id = $11 AND archived_at IS NULL
		   ))
		 RETURNING id`,
		ownerType, req.OwnerId, req.Name, req.Street, sqlutil.NullableStr(req.AdditionalStreet),
		req.City, req.ZipCode, req.CountryId, sqlutil.NullableStr(req.Email), sqlutil.NullableStr(req.Phone),
		req.AuthUserId,
	).Scan(&addressID)
	if queryErr == sql.ErrNoRows {
		return &usersGrpc.CreateAddressResponse{Success: false, Code: codes.NotFound}, nil
	}
	if queryErr != nil {
		return &usersGrpc.CreateAddressResponse{Success: false, Code: codes.InternalError}, queryErr
	}

	return &usersGrpc.CreateAddressResponse{Success: true, Code: codes.Success, AddressId: addressID}, nil
}
