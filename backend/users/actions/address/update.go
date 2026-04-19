package address

import (
	"context"
	"database/sql"

	usersGrpc "project-devis-users/services/grpc"
)

func Update(ctx context.Context, db *sql.DB, req *usersGrpc.UpdateAddressRequest) (*usersGrpc.UpdateAddressResponse, error) {
	if req.AddressId == 0 || req.UserId == "" || req.Name == "" || req.Street == "" || req.City == "" || req.ZipCode == "" || req.CountryId == 0 {
		return &usersGrpc.UpdateAddressResponse{Success: false, Code: codeInvalidInput}, nil
	}

	res, err := db.ExecContext(ctx,
		`UPDATE addresses SET name=$1, street=$2, additional_street=$3, city=$4, zip_code=$5,
		        country_id=$6, email=$7, phone=$8, updated_at=NOW()
		 WHERE id=$9 AND user_id=$10 AND archived_at IS NULL`,
		req.Name, req.Street, nullableStr(req.AdditionalStreet), req.City, req.ZipCode,
		req.CountryId, nullableStr(req.Email), nullableStr(req.Phone),
		req.AddressId, req.UserId,
	)
	if err != nil {
		return &usersGrpc.UpdateAddressResponse{Success: false, Code: codeInternalError}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &usersGrpc.UpdateAddressResponse{Success: false, Code: codeNotFound}, nil
	}

	return &usersGrpc.UpdateAddressResponse{Success: true, Code: codeSuccess}, nil
}
