package client

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"project-devis-users/actions/codes"
	"project-devis-users/actions/sqlutil"
	usersGrpc "project-devis-users/services/grpc"
)

func Create(ctx context.Context, db *sql.DB, req *usersGrpc.CreateClientRequest) (*usersGrpc.CreateClientResponse, error) {
	if req.UserId == "" || req.FirstName == "" || req.LastName == "" {
		return &usersGrpc.CreateClientResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	clientID := uuid.New().String()
	_, err := db.ExecContext(ctx,
		`INSERT INTO clients (client_id, user_id, first_name, last_name, email, phone, company, siren, vat)
		 VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9)`,
		clientID, req.UserId, req.FirstName, req.LastName,
		sqlutil.NullableStr(req.Email), sqlutil.NullableStr(req.Phone),
		sqlutil.NullableStr(req.Company), sqlutil.NullableStr(req.Siren), sqlutil.NullableStr(req.Vat),
	)
	if err != nil {
		return &usersGrpc.CreateClientResponse{Success: false, Code: codes.InternalError}, err
	}

	return &usersGrpc.CreateClientResponse{Success: true, Code: codes.Success, ClientId: clientID}, nil
}
