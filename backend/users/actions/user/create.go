package user

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"project-devis-users/actions/codes"
	usersGrpc "project-devis-users/services/grpc"
)

func Create(ctx context.Context, db *sql.DB, req *usersGrpc.CreateUserRequest) (*usersGrpc.CreateUserResponse, error) {
	if req.Email == "" {
		return &usersGrpc.CreateUserResponse{
			Success:          false,
			Code:             codes.InvalidInput,
			ValidationErrors: []*usersGrpc.ValidationError{{Field: "email", Message: "Champ requis."}},
		}, nil
	}

	userID := uuid.New().String()
	_, err := db.ExecContext(ctx,
		"INSERT INTO users (user_id, email) VALUES ($1, $2)",
		userID, req.Email,
	)
	if err != nil {
		if pqErr, ok := err.(*pq.Error); ok && pqErr.Code == "23505" {
			return &usersGrpc.CreateUserResponse{Success: false, Code: codes.AlreadyExists}, nil
		}
		return &usersGrpc.CreateUserResponse{Success: false, Code: codes.InternalError}, err
	}

	return &usersGrpc.CreateUserResponse{Success: true, Code: codes.Success, UserId: userID}, nil
}
