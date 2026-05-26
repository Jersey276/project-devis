package client

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	usersGrpc "project-devis-users/services/grpc"
)

func List(ctx context.Context, db *sql.DB, req *usersGrpc.ListClientsRequest) (*usersGrpc.ListClientsResponse, error) {
	if req.UserId == "" {
		return &usersGrpc.ListClientsResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	query := `SELECT client_id, user_id, first_name, last_name, email, phone, company, siren, vat,
	                 (archived_at IS NOT NULL)
	          FROM clients WHERE user_id=$1`
	if !req.IncludeArchived {
		query += ` AND archived_at IS NULL`
	}
	query += ` ORDER BY id`

	rows, err := db.QueryContext(ctx, query, req.UserId)
	if err != nil {
		return &usersGrpc.ListClientsResponse{Success: false, Code: codes.InternalError}, err
	}
	defer rows.Close()

	var clients []*usersGrpc.Client
	for rows.Next() {
		var c usersGrpc.Client
		var email, phone, company, siren, vat sql.NullString
		if err := rows.Scan(&c.ClientId, &c.UserId, &c.FirstName, &c.LastName, &email, &phone, &company, &siren, &vat, &c.Archived); err != nil {
			return &usersGrpc.ListClientsResponse{Success: false, Code: codes.InternalError}, err
		}
		c.Email = email.String
		c.Phone = phone.String
		c.Company = company.String
		c.Siren = siren.String
		c.Vat = vat.String
		clients = append(clients, &c)
	}
	if err := rows.Err(); err != nil {
		return &usersGrpc.ListClientsResponse{Success: false, Code: codes.InternalError}, err
	}

	return &usersGrpc.ListClientsResponse{Success: true, Code: codes.Success, Clients: clients}, nil
}
