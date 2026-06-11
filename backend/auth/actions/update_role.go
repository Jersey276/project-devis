package actions

import (
	"context"

	authGrpc "project-devis-auth/services/grpc"
)

var validRoles = map[string]bool{
	"free_user":   true,
	"super_admin": true,
}

func (s *Server) UpdateRole(ctx context.Context, req *authGrpc.UpdateRoleRequest) (*authGrpc.GenericResponse, error) {
	if !validRoles[req.GetRole()] {
		return &authGrpc.GenericResponse{Success: false, Code: CodeInvalidInput}, nil
	}

	result, err := s.db.ExecContext(ctx,
		"UPDATE auth SET role = $1, session_version = session_version + 1 WHERE user_id = $2",
		req.GetRole(),
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
