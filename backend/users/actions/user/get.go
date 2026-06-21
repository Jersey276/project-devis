package user

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	usersGrpc "project-devis-users/services/grpc"
)

func Get(ctx context.Context, db *sql.DB, req *usersGrpc.GetUserRequest) (*usersGrpc.GetUserResponse, error) {
	if req.UserId == "" {
		return &usersGrpc.GetUserResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	var u usersGrpc.User
	err := db.QueryRowContext(ctx,
		"SELECT user_id, email, COALESCE(phone,''), COALESCE(company,''), COALESCE(siren,''), COALESCE(vat,''), COALESCE(logo_url,''), suspended, oss_enabled, COALESCE(iban,''), COALESCE(bic,'') FROM users WHERE user_id = $1",
		req.UserId,
	).Scan(&u.UserId, &u.Email, &u.Phone, &u.Company, &u.Siren, &u.Vat, &u.LogoUrl, &u.Suspended, &u.OssEnabled, &u.Iban, &u.Bic)
	if err == sql.ErrNoRows {
		return &usersGrpc.GetUserResponse{Success: false, Code: codes.NotFound}, nil
	}
	if err != nil {
		return &usersGrpc.GetUserResponse{Success: false, Code: codes.InternalError}, err
	}

	return &usersGrpc.GetUserResponse{Success: true, Code: codes.Success, User: &u}, nil
}
