package client

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"project-devis-users/actions/codes"
	"project-devis-users/actions/sqlutil"
	usersGrpc "project-devis-users/services/grpc"
)

func List(ctx context.Context, db *sql.DB, req *usersGrpc.ListClientsRequest) (*usersGrpc.ListClientsResponse, error) {
	if req.UserId == "" {
		return &usersGrpc.ListClientsResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	page := req.Page
	if page < 1 {
		page = 1
	}
	pageSize := req.PageSize
	if pageSize < 1 || pageSize > 200 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	where, args := buildClientFilters(req.UserId, req.IncludeArchived, req.Filters)

	var total int64
	if err := db.QueryRowContext(ctx, "SELECT COUNT(*) FROM clients"+where, args...).Scan(&total); err != nil {
		return &usersGrpc.ListClientsResponse{Success: false, Code: codes.InternalError}, err
	}

	args = append(args, pageSize, offset)
	n := len(args)
	query := fmt.Sprintf(
		`SELECT client_id, user_id, first_name, last_name, email, phone, company, siren, vat, siret, client_type,
		        (archived_at IS NOT NULL), linked_user_id
		 FROM clients%s ORDER BY id LIMIT $%d OFFSET $%d`,
		where, n-1, n,
	)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return &usersGrpc.ListClientsResponse{Success: false, Code: codes.InternalError}, err
	}
	defer rows.Close()

	var clients []*usersGrpc.Client
	for rows.Next() {
		var c usersGrpc.Client
		var email, phone, company, siren, vat, siret, clientType, linkedUserID sql.NullString
		if err := rows.Scan(&c.ClientId, &c.UserId, &c.FirstName, &c.LastName, &email, &phone, &company, &siren, &vat, &siret, &clientType, &c.Archived, &linkedUserID); err != nil {
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

	return &usersGrpc.ListClientsResponse{Success: true, Code: codes.Success, Clients: clients, Total: total}, nil
}

func buildClientFilters(userID string, includeArchived bool, f *usersGrpc.ClientFilters) (string, []interface{}) {
	args := []interface{}{userID}
	clauses := []string{"user_id = $1"}

	if !includeArchived {
		clauses = append(clauses, "archived_at IS NULL")
	}

	if f != nil {
		if f.Search != "" {
			args = append(args, "%"+f.Search+"%")
			n := len(args)
			clauses = append(clauses, fmt.Sprintf(
				"(first_name ILIKE $%d OR last_name ILIKE $%d OR email ILIKE $%d OR company ILIKE $%d)",
				n, n, n, n,
			))
		}
		if len(f.ClientTypes) > 0 {
			placeholders := make([]string, len(f.ClientTypes))
			for i, ct := range f.ClientTypes {
				args = append(args, ct)
				placeholders[i] = fmt.Sprintf("$%d", len(args))
			}
			clauses = append(clauses, "client_type IN ("+strings.Join(placeholders, ",")+")")
		}
	}

	return " WHERE " + strings.Join(clauses, " AND "), args
}
