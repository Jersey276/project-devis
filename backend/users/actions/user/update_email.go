package user

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	usersGrpc "project-devis-users/services/grpc"
)

func UpdateEmail(ctx context.Context, db *sql.DB, req *usersGrpc.UpdateUserEmailRequest) (*usersGrpc.GenericResponse, error) {
	if req.UserId == "" || req.NewEmail == "" {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	var existing string
	err := db.QueryRowContext(ctx, "SELECT user_id FROM users WHERE email = $1 AND user_id != $2", req.NewEmail, req.UserId).Scan(&existing)
	if err == nil {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.AlreadyExists}, nil
	}
	if err != sql.ErrNoRows {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}

	res, err := db.ExecContext(ctx,
		"UPDATE users SET email = $1, updated_at = NOW() WHERE user_id = $2",
		req.NewEmail, req.UserId,
	)
	if err != nil {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.NotFound}, nil
	}

	return &usersGrpc.GenericResponse{Success: true, Code: codes.Success}, nil
}
