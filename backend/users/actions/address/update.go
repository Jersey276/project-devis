package address

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	"project-devis-users/actions/sqlutil"
	usersGrpc "project-devis-users/services/grpc"
)

func Update(ctx context.Context, db *sql.DB, req *usersGrpc.UpdateAddressRequest) (*usersGrpc.UpdateAddressResponse, error) {
	ownerType, err := sqlutil.OwnerTypeToDBString(req.OwnerType)
	if err != nil || req.AddressId == 0 || req.OwnerId == "" || req.AuthUserId == "" ||
		req.Name == "" || req.Street == "" || req.City == "" || req.ZipCode == "" || req.CountryId == 0 {
		return &usersGrpc.UpdateAddressResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	res, execErr := db.ExecContext(ctx,
		`UPDATE addresses SET name=$1, street=$2, additional_street=$3, city=$4, zip_code=$5,
		        country_id=$6, email=$7, phone=$8, updated_at=NOW()
		 WHERE id=$9 AND owner_type=$10 AND owner_id=$11 AND archived_at IS NULL
		   AND `+sqlutil.AddressAuthPredicate(12),
		req.Name, req.Street, sqlutil.NullableStr(req.AdditionalStreet), req.City, req.ZipCode,
		req.CountryId, sqlutil.NullableStr(req.Email), sqlutil.NullableStr(req.Phone),
		req.AddressId, ownerType, req.OwnerId, req.AuthUserId,
	)
	if execErr != nil {
		return &usersGrpc.UpdateAddressResponse{Success: false, Code: codes.InternalError}, execErr
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &usersGrpc.UpdateAddressResponse{Success: false, Code: codes.NotFound}, nil
	}

	return &usersGrpc.UpdateAddressResponse{Success: true, Code: codes.Success}, nil
}
