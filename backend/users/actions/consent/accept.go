package consent

import (
	"context"
	"database/sql"
	"net"
	"strings"

	"project-devis-users/actions/codes"
	usersGrpc "project-devis-users/services/grpc"
)

func Accept(ctx context.Context, db *sql.DB, req *usersGrpc.AcceptConsentRequest) (*usersGrpc.GenericResponse, error) {
	userID := strings.TrimSpace(req.UserId)
	consentType := strings.TrimSpace(req.Type)
	version := strings.TrimSpace(req.Version)

	if userID == "" || consentType == "" || version == "" {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}
	if consentType != "cgv" && consentType != "privacy_policy" && consentType != "cookies" {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InvalidInput}, nil
	}

	var ipVal interface{}
	if ip := net.ParseIP(strings.TrimSpace(req.Ip)); ip != nil {
		ipVal = ip.String()
	}

	_, err := db.ExecContext(ctx,
		`INSERT INTO consents (user_id, type, version, ip) VALUES ($1, $2, $3, $4)`,
		userID, consentType, version, ipVal,
	)
	if err != nil {
		return &usersGrpc.GenericResponse{Success: false, Code: codes.InternalError}, nil
	}

	return &usersGrpc.GenericResponse{Success: true, Code: codes.Success}, nil
}
