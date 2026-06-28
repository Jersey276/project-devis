package user

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	"project-devis-users/actions/sqlutil"
	usersGrpc "project-devis-users/services/grpc"
)

func Update(ctx context.Context, db *sql.DB, req *usersGrpc.UpdateUserRequest) (*usersGrpc.UpdateUserResponse, error) {
	if req.UserId == "" {
		return &usersGrpc.UpdateUserResponse{Success: false, Code: codes.InvalidInput}, nil
	}
	// Surface a bad SIRET on the field, like UpdateClient — a bare InvalidInput
	// reaches the front as a 400 with no clue which field is wrong.
	if msg := sqlutil.ValidateSIRET(req.Siret, req.Siren); msg != "" {
		return &usersGrpc.UpdateUserResponse{
			Success:          false,
			Code:             codes.InvalidInput,
			ValidationErrors: []*usersGrpc.ValidationError{{Field: "siret", Message: msg}},
		}, nil
	}

	res, err := db.ExecContext(ctx,
		`UPDATE users SET phone=$1, company=$2, siren=$3, vat=$4, oss_enabled=$5, iban=$6, bic=$7, siret=$8, first_name=$9, last_name=$10, updated_at=NOW() WHERE user_id=$11`,
		sqlutil.NullableStr(req.Phone), sqlutil.NullableStr(req.Company), sqlutil.NullableStr(req.Siren), sqlutil.NullableStr(req.Vat), req.OssEnabled, sqlutil.NullableStr(req.Iban), sqlutil.NullableStr(req.Bic), sqlutil.NullableStr(sqlutil.NormalizeSIRET(req.Siret)), sqlutil.NullableStr(req.FirstName), sqlutil.NullableStr(req.LastName), req.UserId,
	)
	if err != nil {
		return &usersGrpc.UpdateUserResponse{Success: false, Code: codes.InternalError}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &usersGrpc.UpdateUserResponse{Success: false, Code: codes.NotFound}, nil
	}

	return &usersGrpc.UpdateUserResponse{Success: true, Code: codes.Success}, nil
}
