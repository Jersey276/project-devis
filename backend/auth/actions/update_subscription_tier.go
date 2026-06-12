package actions

import (
	"context"

	authGrpc "project-devis-auth/services/grpc"
)

var validTiers = map[string]bool{
	"free":       true,
	"pro":        true,
	"enterprise": true,
}

func (s *Server) UpdateSubscriptionTier(ctx context.Context, req *authGrpc.UpdateSubscriptionTierRequest) (*authGrpc.GenericResponse, error) {
	if !validTiers[req.GetTier()] {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInvalidInput}, nil
	}

	result, err := s.db.ExecContext(ctx,
		"UPDATE auth SET subscription_tier = $1, session_version = session_version + 1 WHERE user_id = $2",
		req.GetTier(),
		req.GetUserId(),
	)
	if err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, nil
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInternalError}, nil
	}
	if rows == 0 {
		return &authGrpc.GenericResponse{Success: false, Code: CodeUserNotFound}, nil
	}

	return &authGrpc.GenericResponse{Success: true, Code: CodeSuccess}, nil
}
