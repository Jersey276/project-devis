package client

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	usersGrpc "project-devis-users/services/grpc"
)

func Get(ctx context.Context, db *sql.DB, req *usersGrpc.GetClientRequest) (*usersGrpc.GetClientResponse, error) {
	if req.ClientId == "" || req.UserId == "" {
		return &usersGrpc.GetClientResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	var c usersGrpc.Client
	var email, phone, company, siren, vat sql.NullString
	// Exclude archived clients: archived rows are read-only via the
	// list-with-include-archived path. Returning them from Get would let
	// callers (e.g. the gateway's owner-resolver) treat archived clients as
	// valid address owners.
	err := db.QueryRowContext(ctx,
		`SELECT client_id, user_id, first_name, last_name, email, phone, company, siren, vat,
		        (archived_at IS NOT NULL)
		 FROM clients WHERE client_id=$1 AND user_id=$2 AND archived_at IS NULL`,
		req.ClientId, req.UserId,
	).Scan(&c.ClientId, &c.UserId, &c.FirstName, &c.LastName, &email, &phone, &company, &siren, &vat, &c.Archived)
	if err == sql.ErrNoRows {
		return &usersGrpc.GetClientResponse{Success: false, Code: codes.NotFound}, nil
	}
	if err != nil {
		return &usersGrpc.GetClientResponse{Success: false, Code: codes.InternalError}, err
	}

	c.Email = email.String
	c.Phone = phone.String
	c.Company = company.String
	c.Siren = siren.String
	c.Vat = vat.String

	return &usersGrpc.GetClientResponse{Success: true, Code: codes.Success, Client: &c}, nil
}
