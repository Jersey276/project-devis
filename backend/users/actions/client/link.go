package client

import (
	"context"
	"database/sql"
	"strings"

	"project-devis-users/actions/codes"
	usersGrpc "project-devis-users/services/grpc"
)

func LinkUser(ctx context.Context, db *sql.DB, req *usersGrpc.LinkClientUserRequest) (*usersGrpc.GenericResponse, error) {
	if req.ClientId == "" || req.ProviderId == "" || req.LinkedUserId == "" {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	// A provider cannot link themselves as their own client.
	if req.ProviderId == req.LinkedUserId {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	// Verify the client belongs to the provider and is not archived.
	var existingLinkedUserID sql.NullString
	err := db.QueryRowContext(ctx,
		`SELECT linked_user_id FROM clients WHERE client_id=$1 AND user_id=$2 AND archived_at IS NULL`,
		req.ClientId, req.ProviderId,
	).Scan(&existingLinkedUserID)
	if err == sql.ErrNoRows {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.NotFound}, nil
	}
	if err != nil {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}

	// Already linked to this specific user (same provider, same client record).
	if existingLinkedUserID.Valid && existingLinkedUserID.String == req.LinkedUserId {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.AlreadyLinked}, nil
	}

	// Atomic update: only set if this client record is not yet linked to anyone.
	// The composite unique index (linked_user_id, user_id) on the table prevents
	// the same user from being linked twice to the same provider.
	result, err := db.ExecContext(ctx,
		`UPDATE clients SET linked_user_id=$1 WHERE client_id=$2 AND user_id=$3 AND linked_user_id IS NULL`,
		req.LinkedUserId, req.ClientId, req.ProviderId,
	)
	if err != nil {
		// Unique constraint violation: this user is already linked to another client of this provider.
		if strings.Contains(err.Error(), "clients_linked_user_provider_key") {
			return &usersGrpc.GenericResponse{Success: false, Code: codes.AlreadyLinked}, nil
		}
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InternalError}, err
	}
	if rows == 0 {
		// Lost the race — another request linked first.
		return &usersGrpc.GenericResponse{Success: false, Code: codes.AlreadyLinked}, nil
	}

	return &usersGrpc.GenericResponse{Success: true, Code: codes.Success}, nil
}
