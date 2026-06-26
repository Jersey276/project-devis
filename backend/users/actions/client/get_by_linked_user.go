package client

import (
	"context"
	"database/sql"

	"project-devis-users/actions/codes"
	"project-devis-users/actions/sqlutil"
	usersGrpc "project-devis-users/services/grpc"
)

func GetByLinkedUser(ctx context.Context, db *sql.DB, req *usersGrpc.GetClientByLinkedUserRequest) (*usersGrpc.ListClientsResponse, error) {
	if req.LinkedUserId == "" {
		return &usersGrpc.ListClientsResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	rows, err := db.QueryContext(ctx,
		`SELECT client_id, user_id, first_name, last_name, email, phone, company, siren, vat, siret, client_type,
		        (archived_at IS NOT NULL), linked_user_id
		 FROM clients WHERE linked_user_id=$1 AND archived_at IS NULL`,
		req.LinkedUserId,
	)
	if err != nil {
		return &usersGrpc.ListClientsResponse{Success: false, Code: codes.InternalError}, err
	}
	defer rows.Close()

	var clients []*usersGrpc.Client
	for rows.Next() {
		var c usersGrpc.Client
		var email, phone, company, siren, vat, siret, clientType, linkedUserID sql.NullString
		if err := rows.Scan(&c.ClientId, &c.UserId, &c.FirstName, &c.LastName,
			&email, &phone, &company, &siren, &vat, &siret, &clientType, &c.Archived, &linkedUserID); err != nil {
			return &usersGrpc.ListClientsResponse{Success: false, Code: codes.InternalError}, err
		}
		c.Email = email.String
		c.Phone = phone.String
		c.Company = company.String
		c.Siren = siren.String
		c.Vat = vat.String
		c.Siret = siret.String
		c.ClientType = sqlutil.ClientTypeFromDBString(clientType.String)
		c.LinkedUserId = linkedUserID.String
		clients = append(clients, &c)
	}
	if err := rows.Err(); err != nil {
		return &usersGrpc.ListClientsResponse{Success: false, Code: codes.InternalError}, err
	}

	return &usersGrpc.ListClientsResponse{Success: true, Code: codes.Success, Clients: clients}, nil
}
