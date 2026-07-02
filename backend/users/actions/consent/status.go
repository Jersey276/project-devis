package consent

import (
	"context"
	"database/sql"
	"strings"

	"project-devis-users/actions/codes"
	usersGrpc "project-devis-users/services/grpc"
)

func GetStatus(ctx context.Context, db *sql.DB, req *usersGrpc.GetConsentStatusRequest) (*usersGrpc.GetConsentStatusResponse, error) {
	userID := strings.TrimSpace(req.UserId)
	if userID == "" {
		return &usersGrpc.GetConsentStatusResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	rows, err := db.QueryContext(ctx,
		`SELECT DISTINCT ON (type) type, version, accepted_at
		 FROM consents
		 WHERE user_id = $1
		 ORDER BY type, accepted_at DESC`,
		userID,
	)
	if err != nil {
		return &usersGrpc.GetConsentStatusResponse{Success: false, Code: codes.InternalError}, nil
	}
	defer rows.Close()

	var entries []*usersGrpc.ConsentEntry
	for rows.Next() {
		var e usersGrpc.ConsentEntry
		if err := rows.Scan(&e.Type, &e.Version, &e.AcceptedAt); err != nil {
			return &usersGrpc.GetConsentStatusResponse{Success: false, Code: codes.InternalError}, nil
		}
		entries = append(entries, &e)
	}

	return &usersGrpc.GetConsentStatusResponse{
		Success:  true,
		Code:     codes.Success,
		Consents: entries,
	}, nil
}
