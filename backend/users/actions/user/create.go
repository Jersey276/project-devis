package user

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	usersGrpc "project-devis-users/services/grpc"
)

func Create(ctx context.Context, db *sql.DB, req *usersGrpc.CreateUserRequest) (*usersGrpc.CreateUserResponse, error) {
	if req.Email == "" {
		return &usersGrpc.CreateUserResponse{Success: false, Code: codeInvalidInput}, nil
	}

	var existing string
	err := db.QueryRowContext(ctx, "SELECT user_id FROM users WHERE email = $1", req.Email).Scan(&existing)
	if err == nil {
		return &usersGrpc.CreateUserResponse{Success: false, Code: codeAlreadyExists}, nil
	}
	if err != sql.ErrNoRows {
		return &usersGrpc.CreateUserResponse{Success: false, Code: codeInternalError}, err
	}

	userID := uuid.New().String()
	_, err = db.ExecContext(ctx,
		"INSERT INTO users (user_id, email) VALUES ($1, $2)",
		userID, req.Email,
	)
	if err != nil {
		return &usersGrpc.CreateUserResponse{Success: false, Code: codeInternalError}, err
	}

	return &usersGrpc.CreateUserResponse{Success: true, Code: codeSuccess, UserId: userID}, nil
}
