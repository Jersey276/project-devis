package user

import (
	"context"
	"database/sql"

	usersGrpc "project-devis-users/services/grpc"
)

func Update(ctx context.Context, db *sql.DB, req *usersGrpc.UpdateUserRequest) (*usersGrpc.UpdateUserResponse, error) {
	if req.UserId == "" {
		return &usersGrpc.UpdateUserResponse{Success: false, Code: codeInvalidInput}, nil
	}

	res, err := db.ExecContext(ctx,
		`UPDATE users SET phone=$1, company=$2, siren=$3, vat=$4, updated_at=NOW() WHERE user_id=$5`,
		nullableStr(req.Phone), nullableStr(req.Company), nullableStr(req.Siren), nullableStr(req.Vat), req.UserId,
	)
	if err != nil {
		return &usersGrpc.UpdateUserResponse{Success: false, Code: codeInternalError}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &usersGrpc.UpdateUserResponse{Success: false, Code: codeNotFound}, nil
	}

	return &usersGrpc.UpdateUserResponse{Success: true, Code: codeSuccess}, nil
}

func nullableStr(s string) interface{} {
	if s == "" {
		return nil
	}
	return s
}
