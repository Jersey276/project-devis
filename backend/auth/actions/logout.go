package actions

import (
	"context"
	"project-devis-auth/services"
	authGrpc "project-devis-auth/services/grpc"
)

func (s *Server) Logout(ctx context.Context, req *authGrpc.LogoutRequest) (*authGrpc.GenericResponse, error) {
	err := services.DeleteRefreshToken(ctx, s.db, req.RefreshToken)
	if err != nil {
		return &authGrpc.GenericResponse{
			Success: false,
			Code:    CodeInternalError,
		}, err
	}

	return &authGrpc.GenericResponse{
		Success: true,
		Code:    CodeSuccess,
	}, nil
}
