package address

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	"project-devis-users/actions/sqlutil"
	usersGrpc "project-devis-users/services/grpc"
)

func Update(ctx context.Context, db *sql.DB, req *usersGrpc.UpdateAddressRequest) (*usersGrpc.UpdateAddressResponse, error) {
	var fieldErrors []*usersGrpc.ValidationError

	ownerType, err := sqlutil.OwnerTypeToDBString(req.OwnerType)
	if err != nil {
		fieldErrors = append(fieldErrors, &usersGrpc.ValidationError{Field: "owner_type", Message: "Type de propriétaire invalide."})
	}
	if req.AddressId == 0 {
		fieldErrors = append(fieldErrors, &usersGrpc.ValidationError{Field: "address_id", Message: "Champ requis."})
	}
	if req.OwnerId == "" {
		fieldErrors = append(fieldErrors, &usersGrpc.ValidationError{Field: "owner_id", Message: "Champ requis."})
	}
	if req.AuthUserId == "" {
		fieldErrors = append(fieldErrors, &usersGrpc.ValidationError{Field: "auth_user_id", Message: "Champ requis."})
	}
	if req.Name == "" {
		fieldErrors = append(fieldErrors, &usersGrpc.ValidationError{Field: "name", Message: "Champ requis."})
	}
	if req.Street == "" {
		fieldErrors = append(fieldErrors, &usersGrpc.ValidationError{Field: "street", Message: "Champ requis."})
	}
	if req.City == "" {
		fieldErrors = append(fieldErrors, &usersGrpc.ValidationError{Field: "city", Message: "Champ requis."})
	}
	if req.ZipCode == "" {
		fieldErrors = append(fieldErrors, &usersGrpc.ValidationError{Field: "zip_code", Message: "Champ requis."})
	}
	if req.CountryId == 0 {
		fieldErrors = append(fieldErrors, &usersGrpc.ValidationError{Field: "country_id", Message: "Champ requis."})
	}

	if len(fieldErrors) > 0 {
		return &usersGrpc.UpdateAddressResponse{Success: false, Code: codes.InvalidInput, ValidationErrors: fieldErrors}, nil
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
