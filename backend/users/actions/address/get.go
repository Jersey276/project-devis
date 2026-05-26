package address

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	"project-devis-users/actions/sqlutil"
	usersGrpc "project-devis-users/services/grpc"
)

// Each address action enforces ownership against the authenticated user via
// the shared SQL predicate sqlutil.AddressAuthPredicate. This replaces the
// gateway's prior GetClient pre-check round-trip and keeps auth in one place
// — the SQL boundary.

// Get intentionally does NOT filter `archived_at IS NULL` on the address row.
// quotes.address_id is NOT NULL with no cross-service FK, so a quote created
// before archival still references the archived address by id; rendering the
// quote needs to read it back. The `archived` flag on the response lets
// callers display "(archived)" if they want, but the row stays accessible.
//
// List/Update/Archive DO filter — those are interactive paths that must not
// surface archived rows or let writes flow to them.
func Get(ctx context.Context, db *sql.DB, req *usersGrpc.GetAddressRequest) (*usersGrpc.GetAddressResponse, error) {
	ownerType, err := sqlutil.OwnerTypeToDBString(req.OwnerType)
	if err != nil || req.AddressId == 0 || req.OwnerId == "" || req.AuthUserId == "" {
		return &usersGrpc.GetAddressResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	var a usersGrpc.Address
	var ownerTypeDB string
	var addlStreet, email, phone sql.NullString
	queryErr := db.QueryRowContext(ctx,
		`SELECT id, owner_type, owner_id, name, street, additional_street, city, zip_code, country_id, email, phone,
		        (archived_at IS NOT NULL)
		 FROM addresses
		 WHERE id=$1 AND owner_type=$2 AND owner_id=$3
		   AND `+sqlutil.AddressAuthPredicate(4),
		req.AddressId, ownerType, req.OwnerId, req.AuthUserId,
	).Scan(&a.Id, &ownerTypeDB, &a.OwnerId, &a.Name, &a.Street, &addlStreet, &a.City, &a.ZipCode, &a.CountryId, &email, &phone, &a.Archived)
	if queryErr == sql.ErrNoRows {
		return &usersGrpc.GetAddressResponse{Success: false, Code: codes.NotFound}, nil
	}
	if queryErr != nil {
		return &usersGrpc.GetAddressResponse{Success: false, Code: codes.InternalError}, queryErr
	}

	a.OwnerType, _ = sqlutil.OwnerTypeFromDBString(ownerTypeDB)
	a.AdditionalStreet = addlStreet.String
	a.Email = email.String
	a.Phone = phone.String

	return &usersGrpc.GetAddressResponse{Success: true, Code: codes.Success, Address: &a}, nil
}
