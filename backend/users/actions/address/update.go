package address

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	"project-devis-users/actions/sqlutil"
	usersGrpc "project-devis-users/services/grpc"
)

func Update(ctx context.Context, db *sql.DB, req *usersGrpc.UpdateAddressRequest) (*usersGrpc.UpdateAddressResponse, error) {
	if req.AddressId == 0 || req.UserId == "" || req.Name == "" || req.Street == "" || req.City == "" || req.ZipCode == "" || req.CountryId == 0 {
		return &usersGrpc.UpdateAddressResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	res, err := db.ExecContext(ctx,
		`UPDATE addresses SET name=$1, street=$2, additional_street=$3, city=$4, zip_code=$5,
		        country_id=$6, email=$7, phone=$8, updated_at=NOW()
		 WHERE id=$9 AND user_id=$10 AND archived_at IS NULL`,
		req.Name, req.Street, sqlutil.NullableStr(req.AdditionalStreet), req.City, req.ZipCode,
		req.CountryId, sqlutil.NullableStr(req.Email), sqlutil.NullableStr(req.Phone),
		req.AddressId, req.UserId,
	)
	if err != nil {
		return &usersGrpc.UpdateAddressResponse{Success: false, Code: codes.InternalError}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &usersGrpc.UpdateAddressResponse{Success: false, Code: codes.NotFound}, nil
	}

	return &usersGrpc.UpdateAddressResponse{Success: true, Code: codes.Success}, nil
}
