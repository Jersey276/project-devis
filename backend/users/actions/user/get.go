package user

import (
	"context"
	"database/sql"

	usersGrpc "project-devis-users/services/grpc"
)

func Get(ctx context.Context, db *sql.DB, req *usersGrpc.GetUserRequest) (*usersGrpc.GetUserResponse, error) {
	if req.UserId == "" {
		return &usersGrpc.GetUserResponse{Success: false, Code: codeInvalidInput}, nil
	}

	var u usersGrpc.User
	err := db.QueryRowContext(ctx,
		"SELECT user_id, email, COALESCE(phone,''), COALESCE(company,''), COALESCE(siren,''), COALESCE(vat,'') FROM users WHERE user_id = $1",
		req.UserId,
	).Scan(&u.UserId, &u.Email, &u.Phone, &u.Company, &u.Siren, &u.Vat)
	if err == sql.ErrNoRows {
		return &usersGrpc.GetUserResponse{Success: false, Code: codeNotFound}, nil
	}
	if err != nil {
		return &usersGrpc.GetUserResponse{Success: false, Code: codeInternalError}, err
	}

	return &usersGrpc.GetUserResponse{Success: true, Code: codeSuccess, User: &u}, nil
}
